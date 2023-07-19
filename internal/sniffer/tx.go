package sniffer

import "github.com/ethereum/go-ethereum/core/types"

type Transaction struct {
	Method    [4]byte
	Contract  bool
	IsPending bool
	*types.Transaction
}

func NewTransaction(tx *types.Transaction, pending bool) *Transaction {
	var (
		isContract bool
		methodID   [4]byte
	)
	if tx.To() == nil || len(tx.Data()) < 4 {
		isContract = true
	}
	if !isContract {
		copy(methodID[:], tx.Data()[:4])
	}
	return &Transaction{
		Method:      methodID,
		Contract:    isContract,
		IsPending:   pending,
		Transaction: tx,
	}
}
