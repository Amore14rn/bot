package handlers

import (
	"context"
	"cryptobot/internal/client"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"time"
)

type BeeInput struct {
	Address    string `json:"addr"`
	PrimaryKey string `json:"pk"`
}

type Swarm struct {
	Bees    []*Bee
	Conn    *client.EthClient
	ChainID *big.Int
}

func NewSwarm(conn *client.EthClient, input []BeeInput) (*Swarm, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	id, err := conn.Client().ChainID(ctx)
	if err != nil {
		return nil, err
	}
	bees := make([]*Bee, 0, len(input))
	swarm := &Swarm{
		Conn:    conn,
		ChainID: id,
	}
	for _, raw := range input {
		if !common.IsHexAddress(raw.Address) {
			return nil, fmt.Errorf("incorrect bee address %s", raw.Address)
		}
		addr := common.HexToAddress(raw.Address)
		pk, pkErr := crypto.HexToECDSA(raw.PrimaryKey)
		if pkErr != nil {
			return nil, pkErr
		}
		next, nextErr := NewBee(addr, pk, swarm)
		if nextErr != nil {
			return nil, nextErr
		}
		bees = append(bees, next)
	}
	swarm.Bees = bees
	return swarm, nil
}

func NewSwarmFromJSON(conn *client.EthClient, data []byte) (*Swarm, error) {
	bees := make([]BeeInput, 0)
	err := json.Unmarshal(data, &bees)
	if err != nil {
		return nil, err
	}

	return NewSwarm(conn, bees)
}
