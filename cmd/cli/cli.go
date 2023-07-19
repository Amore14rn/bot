package main

import (
	"context"
	"cryptobot/internal/client"
	"cryptobot/pkg/erc20"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	go1inch "github.com/jon4hz/go-1inch"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"log"
	"math/big"
	"os"
	"time"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "slippage",
				Aliases: []string{"s"},
				Value:   "50",
				Usage:   "Slippage. 0-50",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "approve",
				Usage: "approve tokens to sale",
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
					defer cancel()

					viper.AddConfigPath("./configs")
					viper.SetConfigName("sell")
					viper.SetConfigType("yaml")
					err := viper.ReadInConfig()
					if err != nil {
						return err
					}
					viper.SetConfigName("config")
					err = viper.MergeInConfig()
					if err != nil {
						return err
					}

					log.Println("Node URL:", viper.GetString("nodeUrl"))
					conn, err := client.DialWithContext(ctx, viper.GetString("nodeUrl"))
					if err != nil {
						return err
					}

					iClient := go1inch.NewClient()
					health, _, err := iClient.Healthcheck(ctx, "bsc")
					if err != nil {
						return err
					}
					fmt.Println("Health:", health)

					tokenName := c.Args().Get(0)
					cfg := viper.GetStringMapString(tokenName)
					if _, ok := cfg["from"]; !ok {
						return fmt.Errorf("from not found for %s", tokenName)
					}
					if _, ok := cfg["to"]; !ok {
						return fmt.Errorf("to not found for %s", tokenName)
					}

					wallet := common.HexToAddress(viper.GetString("wallet"))
					/*erc, err := erc20.NewErc20(common.HexToAddress(cfg["from"]), conn.Client())
					if err != nil {
						return err
					}
					amount, err := erc.BalanceOf(nil, wallet)
					if err != nil {
						return err
					}*/

					pk, err := crypto.HexToECDSA(viper.GetString("pk"))
					if err != nil {
						return err
					}

					approveRes, _, err := iClient.ApproveTransaction(ctx, "bsc", cfg["from"], &go1inch.ApproveTransactionOpts{})
					if err != nil {
						return err
					}
					nonce, err := conn.Client().PendingNonceAt(ctx, wallet)
					if err != nil {
						return err
					}

					gasPrice := new(big.Int)
					//gasPrice.SetString(approveRes.GasPrice, 16)
					gasPrice.SetInt64(5000000000)
					to := common.HexToAddress(cfg["from"])
					approveTX := types.NewTx(&types.LegacyTx{
						Nonce:    nonce,
						GasPrice: gasPrice,
						Gas:      uint64(500000),
						To:       &to,
						Value:    big.NewInt(0),
						Data:     common.FromHex(approveRes.Data),
					})
					signedApprove, err := types.SignTx(approveTX, types.NewEIP155Signer(big.NewInt(56)), pk)
					err = conn.Client().SendTransaction(ctx, signedApprove)
					if err != nil {
						return err
					}

					fmt.Println(signedApprove.Hash())
					return nil
				},
			},
			{
				Name:  "sell",
				Usage: "sell tokens",
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
					defer cancel()

					viper.AddConfigPath("./configs")
					viper.SetConfigName("sell")
					viper.SetConfigType("yaml")
					err := viper.ReadInConfig()
					if err != nil {
						return err
					}
					viper.SetConfigName("config")
					err = viper.MergeInConfig()
					if err != nil {
						return err
					}

					log.Println("Node URL:", viper.GetString("nodeUrl"))
					conn, err := client.DialWithContext(ctx, viper.GetString("nodeUrl"))
					if err != nil {
						return err
					}

					iClient := go1inch.NewClient()
					health, _, err := iClient.Healthcheck(ctx, "bsc")
					if err != nil {
						return err
					}
					fmt.Println("Health:", health)

					tokenName := c.Args().Get(0)
					cfg := viper.GetStringMapString(tokenName)
					if _, ok := cfg["from"]; !ok {
						return fmt.Errorf("from not found for %s", tokenName)
					}
					if _, ok := cfg["to"]; !ok {
						return fmt.Errorf("to not found for %s", tokenName)
					}

					wallet := common.HexToAddress(viper.GetString("wallet"))
					erc, err := erc20.NewErc20(common.HexToAddress(cfg["from"]), conn.Client())
					if err != nil {
						return err
					}
					amount, err := erc.BalanceOf(nil, wallet)
					if err != nil {
						return err
					}

					slippage := c.Int64("slippage")
					pk, err := crypto.HexToECDSA(viper.GetString("pk"))
					if err != nil {
						return err
					}

					swapRes, _, err := iClient.Swap(ctx, "bsc", cfg["from"], cfg["to"], amount.String(), viper.GetString("wallet"), slippage, &go1inch.SwapOpts{})
					if err != nil {
						return err
					}
					nonce, err := conn.Client().PendingNonceAt(ctx, wallet)
					if err != nil {
						return err
					}
					gasPrice := new(big.Int)
					//gasPrice.SetString(swapRes.GasPrice, 16)
					gasPrice.SetInt64(5000000000)
					to := common.HexToAddress(swapRes.Tx.To)
					amountVal := new(big.Int)
					amountVal.SetString(swapRes.Tx.Value, 16)
					approveTX := types.NewTx(&types.LegacyTx{
						Nonce:    nonce,
						GasPrice: gasPrice,
						Gas:      uint64(500000),
						To:       &to,
						Value:    amountVal,
						Data:     common.FromHex(swapRes.Tx.Data),
					})

					signed, err := types.SignTx(approveTX, types.NewEIP155Signer(big.NewInt(56)), pk)
					err = conn.Client().SendTransaction(ctx, signed)
					if err != nil {
						return err
					}

					fmt.Println(signed.Hash())
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
