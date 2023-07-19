package main

import (
	"context"
	abiContract "cryptobot/internal/abi"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/signal"
	"time"
)

const (
	snipeURL = "wss://speedy-nodes-nyc.moralis.io/361e9d4f9d47c9e6ecae980d/bsc/mainnet/ws"
	workers  = 100
	router   = "0x10ED43C718714eb63d5aA57B78B54704E256024E"
)

/*
# 0x4b b2 78 f3 - PinkSale finalize
# 0x26 7d d1 02 - DxSale
# 0xf3 05 d7 19 - addLiquidityETH (когда пара в WBNB)
# 0xe8 e3 37 00 - addLiquidity (когда пара в BUSD)
# 0x5c 85 97 4f - setTxLimit (open trading)
# 0x31 df 8a a4 - updateTradingEnabled
# 0x1e ed 1a c8 - setTradingOpen
# 0x8f 70 cc f7 - setTrading
# 0xc9 56 7b f9 - openTrading
# 0x8a 8c 52 3c - enableTrading
# 0x29 32 30 b8 - startTrading
*/
// sudo geth snapshot prune-state --datadir /var/lib/binance/data/ --datadir.ancient /var/lib/binance/ancient

/*
	function setTxLimit(uint256 amount) external onlyOwner {
		require(amount >= _totalSupply / 1000);
		_maxTxAmount = amount.div(100);
	}
*/
var LiquidityMethodIDs = map[[4]byte]struct{}{
	//[...]byte{0xf3, 0x05, 0xd7, 0x19}: {}, // addLiquidityETH
	//[...]byte{0xe8, 0xe3, 0x37, 0x00}: {}, // addLiquidity
	//[...]byte{0x4b, 0xb2, 0x78, 0xf3}: {}, // PinkSale finalize
	//[...]byte{0x26, 0x7d, 0xd1, 0x02}: {}, // DxSale https://bscscan.com/tx/0x11848a3d6fb568bef52c3e0e1fd99015b79ec69ac0545c3a4542963bb8890eb5
	//[...]byte{0x5c, 0x85, 0x97, 0x4f}: {}, // setTxLimit https://bscscan.com/tx/0xa21ba6eb6fc6b7deb8dde775ce13cc62f0e2a745cc49161bd3246067ab6c40df
	[...]byte{0x31, 0xdf, 0x8a, 0x4a}: {}, // updateTradingEnabled
	//[...]byte{0x1e, 0xed, 0x1a, 0xc8}: {}, // setTradingOpen https://bscscan.com/tx/0xc28467b555ee0f76d020d649fc880f7ba85a5f4ee24d203e977823691366d362
	//[...]byte{0x8f, 0x70, 0xcc, 0xf7}: {}, // setTrading https://bscscan.com/tx/0xf2c18cf631a9b3c5158c83dbc2bb83746c548f079f1622aecb96f6a114edf9f2
	//[...]byte{0xc9, 0x56, 0x7b, 0xf9}: {}, // openTrading https://bscscan.com/tx/0x44b83cf3e98d3c916b1ea1cdd745cbb7fe9a44346a28ad719587112ea79de16b
	//[...]byte{0x8a, 0x8c, 0x52, 0x3c}: {}, // enableTrading https://bscscan.com/tx/0x2702826ae1dcd0956be9a4992af86f6c1b09df42d18f78068bbe530e22e0153a
	//[...]byte{0x29, 0x32, 0x30, 0xb8}: {}, // startTrading https://bscscan.com/tx/0xdb87c81943a75c67adf3d9ad0f4cef24d8255ae9c44999e9ecf48bd318f77d95
	//[...]byte{0xc2, 0xe5, 0xec, 0x04}: {}, // setTradingEnabled https://bscscan.com/tx/0xba1158e0ff0d94ec9ffff83fd457ba341bf8f235b4f0f640d879f04107b61088
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "ws",
				Required: true,
				Usage:    "WS URL",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "get and ABI parse",
				Subcommands: []*cli.Command{
					{
						Name:  "tx",
						Usage: "get and dump transaction",
						Action: func(c *cli.Context) error {
							ctx, cancel := context.WithCancel(context.Background())
							defer cancel()

							addr := c.Args().First()
							rpcClient, client, err := connect(ctx, c.String("ws"))
							if err != nil {
								return err
							}
							defer rpcClient.Close()
							defer client.Close()

							hash := common.HexToHash(addr)
							tx, pending, err := client.TransactionByHash(ctx, hash)
							if err != nil {
								return err
							}

							rec, err := client.TransactionReceipt(ctx, hash)
							if err != nil {
								return err
							}
							spew.Dump(rec)
							abIface, err := abiContract.NewABI()
							if err != nil {
								return err
							}

							color.Yellow("%s", hash)
							fmt.Println("Pending:", pending)
							if tx.To() == nil {
								fmt.Println("MethodID: None (contract deploying)")
								return nil
							}
							fmt.Printf("MethodID: %#x\n", tx.Data()[:4])
							d, err := abIface.Parse(tx.Data())
							switch err {
							case nil:
							case abiContract.ErrUnknownMethod:
								fmt.Println("ABI Status:", err.Error())
								jsonData, _ := tx.MarshalJSON()
								fmt.Println(string(jsonData))
								spew.Dump(tx)
								return nil
							default:
								return err
							}
							spew.Dump(d)

							return nil
						},
					},
					{
						Name:  "block",
						Usage: "remove an existing template",
						Action: func(c *cli.Context) error {
							fmt.Println("removed task template: ", c.Args().First())
							return nil
						},
					},
				},
			},
			{
				Name:    "sniff",
				Aliases: []string{"s"},
				Usage:   "Sniffing the resource",
				Subcommands: []*cli.Command{
					{
						Name:  "ld",
						Usage: "Sniffing the liquidity",
						Action: func(c *cli.Context) error {
							ctx, cancel := context.WithCancel(context.Background())
							defer cancel()

							//contract := c.Args().First()
							//spew.Dump(contract)
							rpcClient, client, err := connect(ctx, c.String("ws"))
							if err != nil {
								return err
							}
							defer rpcClient.Close()
							defer client.Close()

							abIface, err := abiContract.NewABI()
							if err != nil {
								return err
							}

							ch := make(chan interface{}, workers)
							sub, err := rpcClient.EthSubscribe(ctx, ch, "newPendingTransactions")
							if err != nil {
								return err
							}

							for i := 0; i < workers; i++ {
								go func() {
									for {
										select {
										case val := <-ch:
											hash := common.HexToHash(val.(string))
											//fmt.Println("New tx", hash)
											tx, pending, cErr := client.TransactionByHash(ctx, hash)
											if cErr == ethereum.NotFound {
												//fmt.Println("Not found", hash)
												continue
											}
											if cErr != nil {
												log.Println(cErr)
												continue
											}
											if tx.To() == nil || len(tx.Data()) < 4 {
												continue
											}

											methodID := [4]byte{}
											copy(methodID[:], tx.Data()[:4])

											if _, ok := LiquidityMethodIDs[methodID]; ok {
												data, pErr := abIface.Parse(tx.Data())
												switch pErr {
												case nil:
												case abiContract.ErrUnknownMethod:
													fmt.Println("ABI Status:", pErr.Error())
												default:
													fmt.Println(pErr)
												}

												color.Yellow("%s", hash)
												fmt.Println("Pending:", pending)
												fmt.Printf("MethodID: %#x\n", tx.Data()[:4])
												spew.Dump(data)
												/*hash1 := common.HexToHash(contract)
												if common.HexToHash(data["token"].(string)) == hash1 || common.HexToHash(data["tokenA"].(string)) == hash1 || common.HexToHash(data["tokenB"].(string)) == hash1 {
													fmt.Println("!!!FOUND")
													ctx.Done()
												}*/
											}

										case <-ctx.Done():
											return
										}
									}
								}()
							}

							signalCh := make(chan os.Signal, 1)
							signal.Notify(signalCh, os.Interrupt)
							go func() {
								call := <-signalCh
								log.Printf("system call: %+v", call)
								cancel()
							}()

							<-ctx.Done()
							sub.Unsubscribe()
							close(ch)
							time.Sleep(time.Second * 1)
							return nil
						},
					},
					{
						Name:  "ld",
						Usage: "Sniffing the liquidity",
						Action: func(c *cli.Context) error {
							ctx, cancel := context.WithCancel(context.Background())
							defer cancel()

							contract := c.Args().First()
							rpcClient, client, err := connect(ctx, c.String("ws"))
							if err != nil {
								return err
							}
							defer rpcClient.Close()
							defer client.Close()

							abIface, err := abiContract.NewABI()
							if err != nil {
								return err
							}

							ch := make(chan interface{}, workers)
							sub, err := rpcClient.EthSubscribe(ctx, ch, "newPendingTransactions")
							if err != nil {
								return err
							}

							for i := 0; i < workers; i++ {
								go func() {
									for {
										select {
										case val := <-ch:
											hash := common.HexToHash(val.(string))
											//fmt.Println("New tx", hash)
											tx, pending, cErr := client.TransactionByHash(ctx, hash)
											if cErr == ethereum.NotFound {
												fmt.Println("Not found", hash)
												continue
											}
											if cErr != nil {
												log.Println(cErr)
												continue
											}
											if tx.To() == nil || len(tx.Data()) < 4 {
												continue
											}

											methodID := [4]byte{}
											copy(methodID[:], tx.Data()[:4])

											if _, ok := LiquidityMethodIDs[methodID]; ok {
												data, pErr := abIface.Parse(tx.Data())
												switch pErr {
												case nil:
												case abiContract.ErrUnknownMethod:
													fmt.Println("ABI Status:", pErr.Error())
												default:
													fmt.Println(pErr)
												}

												color.Yellow("%s", hash)
												fmt.Println("Pending:", pending)
												fmt.Printf("MethodID: %#x\n", tx.Data()[:4])
												spew.Dump(data)

												if data["token"] == contract || data["tokenA"] == contract || data["tokenB"] == contract {
													fmt.Println("!!!FOUND")
												}
											}

										case <-ctx.Done():
											return
										}
									}
								}()
							}

							signalCh := make(chan os.Signal, 1)
							signal.Notify(signalCh, os.Interrupt)
							go func() {
								call := <-signalCh
								log.Printf("system call: %+v", call)
								cancel()
							}()

							<-ctx.Done()
							sub.Unsubscribe()
							close(ch)
							time.Sleep(time.Second * 1)
							return nil
						},
					},
					{
						Name:  "block",
						Usage: "remove an existing template",
						Action: func(c *cli.Context) error {
							fmt.Println("removed task template: ", c.Args().First())
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func connect(ctx context.Context, url string) (*rpc.Client, *ethclient.Client, error) {
	rpcClient, err := rpc.DialContext(ctx, url)
	if err != nil {
		return nil, nil, err
	}
	ethClient := ethclient.NewClient(rpcClient)
	return rpcClient, ethClient, nil
}
