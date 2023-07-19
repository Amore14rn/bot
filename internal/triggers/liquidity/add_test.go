package liquidity

import (
	"context"
	"cryptobot/internal/client"
	"cryptobot/internal/sniffer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

const AddLiquidityTXHash = "0x8a18dcaa6f684b041541dc0c680bbad79722ac0e488eda87e1329b0df519944b"

// TODO: rewrite to Suite
func TestAddLiquidity(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := client.DialWithContext(ctx, "ws://49.12.132.53:8546")
	assert.NoError(t, err)
	defer conn.Close()

	tx, pending, err := conn.GetTX(ctx, common.HexToHash(AddLiquidityTXHash))
	assert.NoError(t, err)
	lqTrigger, err := NewAddLiquidity(conn,
		"0xca830317146BfdDE71e7C0B880e2ec1f66E273EE",
		"0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c",
		big.NewInt(1),
	)
	assert.NoError(t, err)
	result := lqTrigger.Trigger(sniffer.NewTransaction(tx, pending))
	assert.Equal(t, true, result)
}
