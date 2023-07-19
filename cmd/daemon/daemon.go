package main

import (
	"context"
	"crypto/md5"
	"cryptobot/internal/client"
	"cryptobot/internal/handlers"
	"cryptobot/internal/handlers/buyers"
	"cryptobot/internal/sniffer"
	"cryptobot/internal/triggers"
	"cryptobot/internal/triggers/liquidity"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"os/signal"
)

var coinsFileHash = ""

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		call := <-signalCh
		log.Printf("system call: %+v", call)
		cancel()
	}()

	viper.AddConfigPath("./configs")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Node URL:", viper.GetString("nodeUrl"))
	conn, err := client.DialWithContext(ctx, viper.GetString("nodeUrl"))
	if err != nil {
		log.Fatal(err)
	}

	swarmJSON, err := ioutil.ReadFile("./configs/bee.json")
	swarm, err := handlers.NewSwarmFromJSON(conn, swarmJSON)
	if err != nil {
		log.Fatal(err)
	}

	viper.SetConfigName("coins")
	err = viper.MergeInConfig()
	if err != nil {
		log.Fatal(err)
	}
	var coins []Coin
	err = viper.UnmarshalKey("coins", &coins)
	if err != nil {
		log.Fatal(err)
	}

	hs, err := GetHandlers(coins, conn, swarm)
	if err != nil {
		log.Fatal(err)
	}

	sniff, err := sniffer.NewSniffer(conn, sniffer.Options{
		Workers:  3000,
		Handlers: hs,
	})
	if err != nil {
		log.Fatal(err)
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		currentHash := fileHash(e.Name)
		if coinsFileHash != currentHash {
			log.Println("Reloading handlers because coins.yaml is changed")
			err = viper.MergeInConfig()
			if err != nil {
				log.Println(err)
				return
			}
			var newCoins []Coin
			err = viper.UnmarshalKey("coins", &newCoins)
			if err != nil {
				log.Println(err)
				return
			}
			if len(newCoins) == 0 {
				log.Println("There is no coins to get new handlers")
				return
			}
			newHandlers, newHandlersErr := GetHandlers(newCoins, conn, swarm)
			if err != nil {
				log.Println(newHandlersErr)
				return
			}
			sniff.SetHandlers(newHandlers)
			coinsFileHash = currentHash
		}
	})

	if err := sniff.Run(ctx); err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
	close(signalCh)
}

type Coin struct {
	Name         string   `mapstructure:"name"`
	Address      string   `mapstructure:"address"`
	Pair         string   `mapstructure:"pair"`
	Trigger      string   `mapstructure:"trigger"`
	MinAmount    int64    `mapstructure:"minAmount"`
	Scanners     []string `mapstructure:"scanners"`
	PinkSaleAddr string   `mapstructure:"pinkSaleAddr"`
}

// TODO: Refactor return errors
func GetHandlers(coins []Coin, conn *client.EthClient, swarm *handlers.Swarm) ([]sniffer.Handler, error) {
	hs := make([]sniffer.Handler, 0, len(coins))
	for _, coin := range coins {
		if !common.IsHexAddress(coin.Address) {
			log.Println(fmt.Sprintf("invalid address %s for %s", coin.Address, coin.Name))
			continue
		}
		if !common.IsHexAddress(coin.Trigger) {
			log.Println(fmt.Sprintf("invalid trigger %s for %s", coin.Trigger, coin.Name))
			continue
		}
		if !common.IsHexAddress(coin.Pair) {
			log.Println(fmt.Sprintf("invalid pair %s for %s", coin.Pair, coin.Name))
			continue
		}
		if len(coin.Scanners) == 0 {
			log.Println(fmt.Sprintf("no scanners for %s", coin.Name))
			continue
		}
		tgs := make([]triggers.Trigger, 0)
		for _, scannerType := range coin.Scanners {
			switch scannerType {
			case "liquidity":
				add, addErr := liquidity.NewAddLiquidity(conn, coin.Address, coin.Pair, big.NewInt(coin.MinAmount))
				if addErr != nil {
					log.Println(addErr)
					continue
				}
				addETH, addETHErr := liquidity.NewAddETHLiquidity(conn, coin.Address, big.NewInt(coin.MinAmount))
				if addETHErr != nil {
					log.Println(addETHErr)
					continue
				}
				tgs = append(tgs, add, addETH)
			case "pinkSale":
				if !common.IsHexAddress(coin.PinkSaleAddr) {
					log.Println(fmt.Sprintf("invalid pinksale address for %s", coin.Name))
					continue
				}
				sale := liquidity.NewFinalize(coin.PinkSaleAddr)
				tgs = append(tgs, sale)
			case "openTrading":
				//TODO: Add open trading when it would be done
			default:
				log.Println(fmt.Sprintf("unknown scanner %s for %s", scannerType, coin.Name))
				continue
			}
		}
		handler := handlers.NewHandler(tgs, buyers.NewContract(swarm, common.HexToAddress(coin.Trigger)))
		hs = append(hs, handler)
		log.Println(fmt.Sprintf("Handler for %s added with %d triggers and %s types", coin.Name, len(tgs), coin.Scanners))
	}
	return hs, nil
}

func fileHash(fn string) string {
	file, err := os.Open(fn)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Println(err)
	}
	return hex.EncodeToString(hash.Sum(nil))
}
