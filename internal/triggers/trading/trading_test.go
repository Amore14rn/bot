package trading

import (
	"context"
	"cryptobot/internal/client"
	"cryptobot/internal/sniffer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTrading(t *testing.T) {
	hashes := map[string]string{
		"0x5a2dda0beedde43444daed62365c940074fadc2c": "0xc28467b555ee0f76d020d649fc880f7ba85a5f4ee24d203e977823691366d362",
		"0xf93c5222a1dd5f482701d6ad6ab9f5eed4b0718b": "0xf2c18cf631a9b3c5158c83dbc2bb83746c548f079f1622aecb96f6a114edf9f2",
		"0x9ac7c93ce54c1e7273c0208904f922abd7f6d317": "0x44b83cf3e98d3c916b1ea1cdd745cbb7fe9a44346a28ad719587112ea79de16b",
		"0xb44634fc5cb905946a9c2c1c01be54844b3c990e": "0x2702826ae1dcd0956be9a4992af86f6c1b09df42d18f78068bbe530e22e0153a",
		"0x9c1309517b2cb37370afae57963d9cedb28d3961": "0xdb87c81943a75c67adf3d9ad0f4cef24d8255ae9c44999e9ecf48bd318f77d95",
		"0x4b0b206d43d8bd2c52147f46436c62fb4c214388": "0xba1158e0ff0d94ec9ffff83fd457ba341bf8f235b4f0f640d879f04107b61088",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := client.DialWithContext(ctx, "ws://49.12.132.53:8546")
	assert.NoError(t, err)
	defer conn.Close()

	for contract, hash := range hashes {
		tx, pending, err := conn.GetTX(ctx, common.HexToHash(hash))
		assert.NoError(t, err)
		trigger := NewOpenTrading(common.HexToAddress(contract))
		assert.NoError(t, err)
		result := trigger.Trigger(sniffer.NewTransaction(tx, pending))
		assert.Equal(t, true, result)
	}
}
