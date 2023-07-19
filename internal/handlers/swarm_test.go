package handlers

import (
	"bytes"
	"context"
	"cryptobot/internal/client"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"log"
	"math/big"
	"testing"
)

var beesJSON = []byte(`[{"pk":"a990752c6b645e4281befcf3c7d8ddc189a3d5f477076fc0ca2523709a8f6588","addr":"0x9a173141f22b1b5406587f6E3C37e31bC2158261"},{"pk":"458976e5d954a35a6d9d2099c72b528e002190671a365d3020d89c1a61945e4b","addr":"0x3CF85BcA80a66C4639660237c2a0603B42261705"},{"pk":"ac2bbd8762f0d0e28a213e01894529a23566c8e345afe72918a4828fe3eea96a","addr":"0x36710dFaF38d11AB3340b4585D77b9fdbE9EFEaD"},{"pk":"72da338377d0c51da3d811c3977c295fd3b053a96ce169164af325dfb9a2ca6f","addr":"0x28Aaf5B73A7d719296Fd99aedd7bD574d3d8CC7d"},{"pk":"5d33c873916671a2ae3b80615cffc0b254e61acac1bc16bb1f146874fbd15537","addr":"0x00674ec96Db3a1599937C092e9aB653f9091fe4f"},{"pk":"de21facd7df4a9e666ff39d9f4816c4153713f56fffaddf07575a55dec2f45b4","addr":"0xd06559C97bdc9f85520dCd6312adADbE9DfBEdeC"},{"pk":"d98886b1ed1986c1ce1fc6033533a85499e4b4e2744b47ca8f56aec5aa44b206","addr":"0xF546679FBf85EC95439b503de334c6222c4744c1"},{"pk":"b805fc022a09c79883a7386b8d85eaa4eefcb0a8f417ad4474ba907d1a3d7075","addr":"0x6a51a10fCfC6CdEEf0d1E4CFB1B64E0893C3BC4B"},{"pk":"db7d239c23bae30eaa52877ab091c056bd5993d41220ec94ac2d0cf588a220eb","addr":"0x2C3f8B4E095484696C9AAE85124A961B01dc16aE"},{"pk":"9c2cad7b6dfbf2adab69e9bd464e697f7e0535bc59b3b1dd259c0d4eaa51fd6d","addr":"0xedD6710bb8ae39B85929f8a0eAab492352820B71"},{"pk":"ff5d288a259d87e97f33b998ec51bddd96340bcaed8cb891688a05ab491ed173","addr":"0x5fd1144c5B998a4f8101C17e65D15b8A50B4e517"},{"pk":"5792f599fa5c7fdf424c2b08ef827145d9195e2a83eb79fa2c5373f6c0369715","addr":"0xEdC03A34E000e38c08c4B280Ad8D8E6C9b83E540"},{"pk":"8e9b196f35cde99eb07074fb6598418c846e4ef8fbd7811a520cda4a417254ba","addr":"0x81E1D0b7B34E93c846F82E67882018b7Ba536ca5"},{"pk":"7c554275a933c974de8185f5e6cbbc102735379c99a8c1803569e54387de1cf8","addr":"0xCdDC82Fd88C0ADCdC6a48Bb6eff96d6107d2ab29"},{"pk":"296015c01ae88d1021e66efaa7fc9831e9db0b994bd6e3ac3330ef29bb605079","addr":"0x7e6aCF1267E7B8369BCC77fd225dEBc183cDCAe6"},{"pk":"82a68efad4b9eaddd697c760249d22c6e32fb0ad2c7837205491773b8b3f41f4","addr":"0xfdEc5D1A6aAe96ea0F277CBe72D9B4EF91658b44"},{"pk":"a8d8e7b042ba306c8247b4eaf23340d6b34cb0ef76fdd607c7155cee36c02c08","addr":"0x486a0180780908a91aF345Bd1EAE6303532f9eb1"}]`)

func TestSwarm(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := client.DialWithContext(ctx, "ws://49.12.132.53:8546")
	assert.NoError(t, err)
	defer conn.Close()
	swarm, err := NewSwarmFromJSON(conn, beesJSON)
	assert.NoError(t, err)

	addr := common.HexToAddress("0xD0561f5357e716dD4741BC033A8329b1Baf5Db51")
	legacyTx := &types.LegacyTx{
		To:       &addr,
		Value:    big.NewInt(0),
		Gas:      5000,
		GasPrice: big.NewInt(500000),
		Data:     []byte{0x81, 0x19, 0xc0, 0x65},
	}

	buff := bytes.NewBuffer(make([]byte, 512))
	log.SetOutput(buff)

	keys := map[string]struct{}{}
	for _, bee := range swarm.Bees {
		tx, bErr := bee.MakeLegacy(legacyTx)
		assert.NoError(t, bErr)
		if _, ok := keys[tx.Hash().Hex()]; ok {
			fmt.Println(keys)
			t.Fatal("Duplicate tx hash", tx.Hash().Hex())
		}
		keys[tx.Hash().Hex()] = struct{}{}
	}
}
