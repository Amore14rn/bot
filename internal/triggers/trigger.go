package triggers

import "cryptobot/internal/sniffer"

type Trigger interface {
	Trigger(transaction *sniffer.Transaction) bool
}
