package abi

import (
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

var ErrUnknownMethod = errors.New("unknown method")

type ABI struct {
	abi.ABI
}

func NewABI() (*ABI, error) {
	i, err := abi.JSON(strings.NewReader(AbiContract))
	if err != nil {
		return nil, err
	}
	return &ABI{
		ABI: i,
	}, nil
}

func (i *ABI) Parse(data []byte) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	method, err := i.MethodById(data[:4])
	if err != nil {
		if strings.Contains(err.Error(), "no method with id:") {
			return nil, ErrUnknownMethod
		}
		return nil, err
	}
	err = method.Inputs.UnpackIntoMap(result, data[4:])
	if err != nil {
		return nil, err
	}
	return result, nil
}
