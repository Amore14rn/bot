package liquidity

import (
	"context"
	"cryptobot/internal/client"
	"cryptobot/internal/sniffer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

const FinalizeHash = "0x6ca44d95d897977cbde3c94b76a6a0a310b64dfc8d2e938bf8044e1f785b802e"

// TODO: rewrite to Suite
// TODO: make the mock with tx
func TestFinalizeLiquidity(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := client.DialWithContext(ctx, "ws://49.12.132.53:8546")
	assert.NoError(t, err)
	defer conn.Close()

	tx, pending, err := conn.GetTX(ctx, common.HexToHash(FinalizeHash))
	assert.NoError(t, err)
	lqTrigger := NewFinalize("0xfc0a33b9205dc3f7452cb2be4eec110b7808d6b0")
	assert.NoError(t, err)
	result := lqTrigger.Trigger(sniffer.NewTransaction(tx, pending))
	assert.Equal(t, true, result)
}
