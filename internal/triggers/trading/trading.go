package trading

import (
	"cryptobot/internal/sniffer"
	"cryptobot/internal/triggers"
	"github.com/ethereum/go-ethereum/common"
)

var TradingMethods = map[[4]byte]struct{}{
	[...]byte{0x31, 0xdf, 0x8a, 0x4a}: {}, // updateTradingEnabled
	[...]byte{0x1e, 0xed, 0x1a, 0xc8}: {}, // setTradingOpen https://bscscan.com/tx/0xc28467b555ee0f76d020d649fc880f7ba85a5f4ee24d203e977823691366d362
	[...]byte{0x8f, 0x70, 0xcc, 0xf7}: {}, // setTrading https://bscscan.com/tx/0xf2c18cf631a9b3c5158c83dbc2bb83746c548f079f1622aecb96f6a114edf9f2
	[...]byte{0xc9, 0x56, 0x7b, 0xf9}: {}, // openTrading https://bscscan.com/tx/0x44b83cf3e98d3c916b1ea1cdd745cbb7fe9a44346a28ad719587112ea79de16b
	[...]byte{0x8a, 0x8c, 0x52, 0x3c}: {}, // enableTrading https://bscscan.com/tx/0x2702826ae1dcd0956be9a4992af86f6c1b09df42d18f78068bbe530e22e0153a
	[...]byte{0x29, 0x32, 0x30, 0xb8}: {}, // startTrading https://bscscan.com/tx/0xdb87c81943a75c67adf3d9ad0f4cef24d8255ae9c44999e9ecf48bd318f77d95
	[...]byte{0xc2, 0xe5, 0xec, 0x04}: {}, // setTradingEnabled https://bscscan.com/tx/0xba1158e0ff0d94ec9ffff83fd457ba341bf8f235b4f0f640d879f04107b61088
}

type OpenTrading struct {
	Token common.Address
}

func (f OpenTrading) Trigger(tx *sniffer.Transaction) bool {
	if tx.Contract == true {
		return false
	}
	if *tx.To() != f.Token {
		return false
	}
	if _, ok := TradingMethods[tx.Method]; !ok {
		return false
	}

	return true
}

func NewOpenTrading(token common.Address) triggers.Trigger {
	return &OpenTrading{
		Token: token,
	}
}
