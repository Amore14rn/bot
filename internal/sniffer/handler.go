package sniffer

type Handler interface {
	Compare(transaction *Transaction) bool
	Handle(transaction *Transaction) error
}
