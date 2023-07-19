package liquidity

import (
	"cryptobot/internal/client"
	"cryptobot/internal/sniffer"
	"cryptobot/internal/triggers"
	"cryptobot/pkg/erc20"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var MethodAddETHLiquidity = [...]byte{0xf3, 0x05, 0xd7, 0x19}

type AddETHLiquidityTxData struct {
	TokenAddress       common.Address
	To                 common.Address
	AmountTokenDesired *big.Int
	AmountTokenMin     *big.Int
	AmountETHMin       *big.Int
	Deadline           *big.Int
}

type AddETHLiquidity struct {
	Address      common.Address
	Token        *erc20.Erc20
	MinLiquidity *big.Int
}

func (a AddETHLiquidity) Trigger(tx *sniffer.Transaction) bool {
	if tx.Method != MethodAddETHLiquidity {
		return false
	}
	input := AddETHLiquidityParseTx(tx.Data()[4:])

	if input.TokenAddress != a.Address {
		return false
	}

	senderBalance, err := a.Token.BalanceOf(nil, input.To)
	if err != nil {
		fmt.Println(fmt.Sprintf("error getting balance of token to buy: %s", err))
		return false
	}

	compareBalance := input.AmountTokenMin.Cmp(senderBalance)
	if compareBalance != 0 && compareBalance != -1 {
		return false
	}

	if tx.Value().Cmp(a.MinLiquidity) != 1 {
		return false
	}

	if input.AmountETHMin.Cmp(a.MinLiquidity) != 1 {
		return false
	}

	return true
}

// MinLiquidity in token which paired
// This method works only with BNB. It means no need to provide pair
func NewAddETHLiquidity(client *client.EthClient, address string, min *big.Int) (triggers.Trigger, error) {
	trigger := &AddETHLiquidity{}
	trigger.Address = common.HexToAddress(address)
	trigger.MinLiquidity = min

	erc, err := erc20.NewErc20(trigger.Address, client.Client())
	if err != nil {
		return nil, err
	}
	trigger.Token = erc

	return trigger, nil
}

func AddETHLiquidityParseTx(data []byte) AddETHLiquidityTxData {
	token := common.BytesToAddress(data[12:32])
	var amountTokenDesired = new(big.Int)
	amountTokenDesired.SetString(common.Bytes2Hex(data[32:64]), 16)
	var amountTokenMin = new(big.Int)
	amountTokenMin.SetString(common.Bytes2Hex(data[64:96]), 16)
	var amountETHMin = new(big.Int)
	amountETHMin.SetString(common.Bytes2Hex(data[96:128]), 16)

	to := common.BytesToAddress(data[140:160])
	var deadline = new(big.Int)
	deadline.SetString(common.Bytes2Hex(data[160:192]), 16)

	return AddETHLiquidityTxData{
		TokenAddress:       token,
		AmountTokenDesired: amountTokenDesired,
		AmountETHMin:       amountETHMin,
		AmountTokenMin:     amountTokenMin,
		Deadline:           deadline,
		To:                 to,
	}
}
