package liquidity

import (
	"cryptobot/internal/sniffer"
	"cryptobot/internal/triggers"
	"github.com/ethereum/go-ethereum/common"
)

var MethodFinalize = [...]byte{0x4b, 0xb2, 0x78, 0xf3}

type Finalize struct {
	PreSaleAddr common.Hash
}

func (f Finalize) Trigger(tx *sniffer.Transaction) bool {
	if tx.Contract == true {
		return false
	}
	if tx.To().Hash() == f.PreSaleAddr && tx.Method == MethodFinalize {
		return true
	}

	return false
}

func NewFinalize(presale string) triggers.Trigger {
	return &Finalize{
		PreSaleAddr: common.HexToHash(presale),
	}
}
