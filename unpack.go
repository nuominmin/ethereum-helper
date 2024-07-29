package ethereumhelper

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
)

import (
	"errors"
)

var (
	errNoEventSignature       = errors.New("no event signature")
	errEventSignatureMismatch = errors.New("event signature mismatch")
)

func Unpack[T any](parsedABI abi.ABI, name string, log *types.Log) (event *T, err error) {
	event = new(T)
	if len(log.Topics) == 0 {
		return event, errNoEventSignature
	}
	if log.Topics[0] != parsedABI.Events[name].ID {
		return event, errEventSignatureMismatch
	}
	if len(log.Data) > 0 {
		if err = parsedABI.UnpackIntoInterface(event, name, log.Data); err != nil {
			return event, err
		}
	}
	var indexed abi.Arguments
	for _, arg := range parsedABI.Events[name].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return event, abi.ParseTopics(event, indexed, log.Topics[1:])
}
