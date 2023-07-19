package client

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type EthClient struct {
	rpcClient *rpc.Client
	ethClient *ethclient.Client
}

// TODO: Connection timeout
func DialWithContext(ctx context.Context, addr string) (*EthClient, error) {
	rpcClient, err := rpc.DialContext(ctx, addr)
	if err != nil {
		return nil, err
	}

	return &EthClient{
		rpcClient: rpcClient,
		ethClient: ethclient.NewClient(rpcClient),
	}, nil
}

func NewETHClient(client *rpc.Client) *EthClient {
	return &EthClient{
		rpcClient: client,
		ethClient: ethclient.NewClient(client),
	}
}

// TODO: Reconnect when dropped
func (c *EthClient) RPCClient() *rpc.Client {
	return c.rpcClient
}

func (c *EthClient) Client() *ethclient.Client {
	return c.ethClient
}

func (c *EthClient) Close() {
	c.ethClient.Close()
	c.rpcClient.Close()
}

func (c *EthClient) GetTX(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	return c.ethClient.TransactionByHash(ctx, hash)
}
