package sniffer

import (
	"context"
	"cryptobot/internal/client"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"sync"
)

// TODO: Rename to Sniffer
type Sniffer struct {
	client      *client.EthClient
	workers     int
	handlers    []Handler
	sub         *rpc.ClientSubscription
	newHashesCh chan interface{}
	sync.RWMutex
}

type Options struct {
	Handlers []Handler
	Workers  int
}

func NewSniffer(client *client.EthClient, opts Options) (*Sniffer, error) {
	if client == nil {
		return nil, errors.New("client is empty")
	}
	if opts.Workers == 0 {
		return nil, errors.New("too small workers value to start")
	}
	if len(opts.Handlers) == 0 {
		return nil, errors.New("has no one handler to sniff")
	}

	return &Sniffer{
		client:   client,
		handlers: opts.Handlers,
		workers:  opts.Workers,
	}, nil
}

func (s *Sniffer) SetHandlers(h []Handler) {
	s.Lock()
	defer s.Unlock()

	s.handlers = h
}

func (s *Sniffer) Run(ctx context.Context) error {
	var err error
	s.newHashesCh = make(chan interface{}, s.workers)
	s.sub, err = s.client.RPCClient().EthSubscribe(ctx, s.newHashesCh, "newPendingTransactions")
	if err != nil {
		return err
	}

	for i := 0; i < s.workers; i++ {
		go func() {
			for {
				select {
				case val := <-s.newHashesCh:
					str, ok := val.(string)
					if !ok {
						fmt.Println("Can't cast", str)
						continue
					}
					hash := common.HexToHash(str)
					tx, pending, cErr := s.client.Client().TransactionByHash(ctx, hash)
					if cErr == ethereum.NotFound {
						//fmt.Println("Not found", hash)
						continue
					}
					if cErr != nil {
						fmt.Println(cErr)
						continue
					}

					preparedTx := NewTransaction(tx, pending)
					s.RLock()
					for _, handler := range s.handlers {
						if handler.Compare(preparedTx) {
							if hErr := handler.Handle(preparedTx); hErr != nil {
								fmt.Println(err)
							}
						}
					}
					s.RUnlock()
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	return nil
}

func (s *Sniffer) Stop() {
	s.client.Close()
	s.sub.Unsubscribe()
}
