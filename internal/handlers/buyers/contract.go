package buyers

import (
	"cryptobot/internal/handlers"
	"cryptobot/internal/sniffer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"math/big"
)

var (
	ContractMethod     = []byte{0x81, 0x19, 0xc0, 0x65} // swap()
	ContractTxValue    = big.NewInt(0)
	ContractTxGasLimit = uint64(500000)
)

type Contract struct {
	trigger common.Address
	swarm   *handlers.Swarm
}

func (c *Contract) Buy(tx *sniffer.Transaction) error {
	log.Println("Start sniping for", tx.Hash().Hex())

	legacyTx := &types.LegacyTx{
		To:       &c.trigger,
		Value:    ContractTxValue,
		Gas:      ContractTxGasLimit,
		GasPrice: tx.GasPrice(),
		Data:     ContractMethod,
	}

	for _, bee := range c.swarm.Bees {
		go func(b *handlers.Bee, legacyTx types.LegacyTx) {
			/*rand.Seed(time.Now().UnixNano())
			rndInt := rand.Intn(999999 - 100 + 1) + 100
			rndBytes := []byte(strconv.FormatInt(int64(rndInt), 16))
			bytes := make([]byte, 32-len(rndBytes), 32)
			bytes = append(bytes, rndBytes...)
			legacyTx.Data = append(legacyTx.Data, bytes...)*/
			_, err := b.SendTx(&legacyTx)
			if err != nil {
				log.Println(err)
				return
			}
		}(bee, *legacyTx)
	}

	return nil
}

func NewContract(swarm *handlers.Swarm, trigger common.Address) handlers.Buyer {
	return &Contract{
		trigger: trigger,
		swarm:   swarm,
	}
}
