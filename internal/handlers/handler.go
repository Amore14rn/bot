package handlers

import (
	"cryptobot/internal/sniffer"
	"cryptobot/internal/triggers"
)

type Buyer interface {
	Buy(tx *sniffer.Transaction) error
}

type Handler struct {
	triggers []triggers.Trigger
	buyer    Buyer
}

func (h Handler) Compare(tx *sniffer.Transaction) bool {
	for _, trigger := range h.triggers {
		if trigger.Trigger(tx) {
			return true
		}
	}

	return false
}

func (h Handler) Handle(tx *sniffer.Transaction) error {
	return h.buyer.Buy(tx)
}

func NewHandler(triggers []triggers.Trigger, buyer Buyer) sniffer.Handler {
	return &Handler{triggers: triggers, buyer: buyer}
}
