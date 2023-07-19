package liquidity

import (
	"cryptobot/internal/client"
	"cryptobot/internal/sniffer"
	"cryptobot/internal/triggers"
	"cryptobot/pkg/erc20"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

var MethodAddLiquidity = [...]byte{0xe8, 0xe3, 0x37, 0x00}

type AddLiquidityTxData struct {
	TokenAddressA       common.Address
	TokenAddressB       common.Address
	To                  common.Address
	AmountTokenADesired *big.Int
	AmountTokenBDesired *big.Int
	AmountTokenAMin     *big.Int
	AmountTokenBMin     *big.Int
	Deadline            *big.Int
}

type AddLiquidity struct {
	Address      common.Address
	Token        *erc20.Erc20
	Pair         common.Address
	MinLiquidity *big.Int
}

func (a AddLiquidity) Trigger(tx *sniffer.Transaction) bool {
	if tx.Method != MethodAddLiquidity {
		return false
	}
	input := AddLiquidityParseTx(tx.Data()[4:])
	if input.TokenAddressA != a.Address && input.TokenAddressB != a.Address {
		return false
	}
	if input.TokenAddressA != a.Pair && input.TokenAddressB != a.Pair {
		return false
	}

	senderBalance, err := a.Token.BalanceOf(nil, input.To)
	if err != nil {
		fmt.Println(fmt.Sprintf("error getting balance of token to buy: %s", err))
		return false
	}

	var minAmountOfToken *big.Int
	var minAmountOfPair *big.Int
	if input.TokenAddressA == a.Address {
		minAmountOfToken = input.AmountTokenAMin
		minAmountOfPair = input.AmountTokenBMin
	} else {
		minAmountOfToken = input.AmountTokenBMin
		minAmountOfPair = input.AmountTokenAMin
	}

	compareBalance := minAmountOfToken.Cmp(senderBalance)
	if compareBalance != 0 && compareBalance != -1 {
		return false
	}

	if minAmountOfPair.Cmp(a.MinLiquidity) != 1 {
		return false
	}

	return true
}

// MinLiquidity in token which paired
func NewAddLiquidity(client *client.EthClient, address, pair string, min *big.Int) (triggers.Trigger, error) {
	trigger := &AddLiquidity{}
	if !common.IsHexAddress(address) {
		return nil, errors.New("token address is not correct")
	}
	if !common.IsHexAddress(pair) {
		return nil, errors.New("token pair address is not correct")
	}
	trigger.Address = common.HexToAddress(address)
	trigger.Pair = common.HexToAddress(pair)
	trigger.MinLiquidity = min

	erc, err := erc20.NewErc20(trigger.Address, client.Client())
	if err != nil {
		return nil, err
	}
	trigger.Token = erc

	return trigger, nil
}

func AddLiquidityParseTx(data []byte) AddLiquidityTxData {
	tokenA := common.BytesToAddress(data[12:32])
	tokenB := common.BytesToAddress(data[44:64])
	var amountTokenADesired = new(big.Int)
	amountTokenADesired.SetString(common.Bytes2Hex(data[64:96]), 16)
	var amountTokenBDesired = new(big.Int)
	amountTokenBDesired.SetString(common.Bytes2Hex(data[96:128]), 16)
	var amountTokenAMin = new(big.Int)
	amountTokenAMin.SetString(common.Bytes2Hex(data[128:160]), 16)
	var amountTokenBMin = new(big.Int)
	amountTokenBMin.SetString(common.Bytes2Hex(data[160:192]), 16)
	to := common.BytesToAddress(data[204:224])
	var deadline = new(big.Int)
	deadline.SetString(common.Bytes2Hex(data[224:256]), 16)

	return AddLiquidityTxData{
		TokenAddressA:       tokenA,
		TokenAddressB:       tokenB,
		AmountTokenADesired: amountTokenADesired,
		AmountTokenBDesired: amountTokenBDesired,
		AmountTokenAMin:     amountTokenAMin,
		AmountTokenBMin:     amountTokenBMin,
		Deadline:            deadline,
		To:                  to,
	}
}
