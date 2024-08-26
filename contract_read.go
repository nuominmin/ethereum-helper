package ethereumhelper

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"io"
	"log"
)

type Contract struct {
	ethClient    *ethclient.Client
	parsedABI    abi.ABI
	contractAddr common.Address
}

func NewContract(ethClient *ethclient.Client, contractAddr string, abiReader io.Reader) *Contract {
	// 解析合约ABI
	parsedABI, err := abi.JSON(abiReader)
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	return &Contract{
		ethClient:    ethClient,
		parsedABI:    parsedABI,
		contractAddr: common.HexToAddress(contractAddr),
	}
}

func ContractRead[T any](ctx context.Context, c *Contract, methodName string) (T, error) {
	var outputData T

	// 准备方法调用
	callData, err := c.parsedABI.Pack(methodName)
	if err != nil {
		return outputData, fmt.Errorf("failed to parse abi, error: %v", err)
	}

	// 执行调用
	var result []byte
	if result, err = c.ethClient.CallContract(ctx, ethereum.CallMsg{
		To:   &c.contractAddr,
		Data: callData,
	}, nil); err != nil {
		return outputData, fmt.Errorf("failed to call contract, error: %v", err)
	}

	// 解析返回值
	if err = c.parsedABI.UnpackIntoInterface(&outputData, methodName, result); err != nil {
		return outputData, fmt.Errorf("failed to unpack result, error: %v", err)
	}

	return outputData, nil
}
