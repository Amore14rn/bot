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

const AddETHLiquidityTXHash = "0x0f55dc5fd06dcca2a0ed44b84ef306fde30f788723946789d6ba032fb65c78bd"

// TODO: rewrite to Suite
// TODO: make the mock with tx
func TestAddETHLiquidity(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := client.DialWithContext(ctx, "ws://49.12.132.53:8546")
	assert.NoError(t, err)
	defer conn.Close()

	tx, pending, err := conn.GetTX(ctx, common.HexToHash(AddETHLiquidityTXHash))
	assert.NoError(t, err)
	lqTrigger, err := NewAddETHLiquidity(conn,
		"0xF0cfD92a363050407468B17648C0677DaF40b351",
		big.NewInt(1),
	)
	assert.NoError(t, err)
	result := lqTrigger.Trigger(sniffer.NewTransaction(tx, pending))
	assert.Equal(t, true, result)
}
