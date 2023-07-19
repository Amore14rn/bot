package handlers

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"time"
)

type Bee struct {
	swarm      *Swarm
	nonce      uint64
	Address    common.Address
	PrimaryKey *ecdsa.PrivateKey
}

func NewBee(addr common.Address, pk *ecdsa.PrivateKey, swarm *Swarm) (*Bee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	nonce, err := swarm.Conn.Client().PendingNonceAt(ctx, addr)
	if err != nil {
		return nil, err
	}
	return &Bee{
		PrimaryKey: pk,
		Address:    addr,
		nonce:      nonce,
		swarm:      swarm,
	}, nil
}

func (b *Bee) SendTx(legacyTx *types.LegacyTx) (common.Hash, error) {
	signed, err := b.MakeLegacy(legacyTx)
	if err != nil {
		return [32]byte{}, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = b.swarm.Conn.Client().SendTransaction(ctx, signed)
	if err != nil {
		log.Println("Error sending tx: ", err.Error(), signed.Hash(), b.Address)
		return common.Hash{}, err
	}

	log.Println("TX sent: ", signed.Hash().Hex(), b.Address)
	b.nonce++

	return signed.Hash(), nil
}

func (b *Bee) MakeLegacy(legacyTx *types.LegacyTx) (*types.Transaction, error) {
	legacyTx.Nonce = b.nonce
	tx := types.NewTx(legacyTx)
	signed, err := types.SignTx(tx, types.NewEIP155Signer(b.swarm.ChainID), b.PrimaryKey)
	if err != nil {
		log.Println("Problem with signed: ", err.Error())
		return signed, err
	}
	log.Println(fmt.Sprintf("Signed: %s", signed.Hash().Hex()))
	return signed, nil
}
