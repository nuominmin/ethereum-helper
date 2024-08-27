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
	"time"
)

type Contract struct {
	ethClient    *ethclient.Client // 以太坊客户端
	parsedABI    abi.ABI           // 合约ABI
	contractAddr common.Address    // 合约地址
	retryNum     int               // 重试次数
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

func (c *Contract) SetRetryNum(n int) {
	c.retryNum = n + 1
}

func ContractRead[T any](ctx context.Context, c *Contract, methodName string, args ...interface{}) (T, error) {
	var outputData T

	// 准备方法调用
	callData, err := c.parsedABI.Pack(methodName, args...)
	if err != nil {
		return outputData, fmt.Errorf("failed to parse abi, error: %v", err)
	}

	num := 1
	if c.retryNum > 0 {
		num += c.retryNum
	}

	// 执行调用
	var result []byte
	for i := 0; i < num; i++ {
		result, err = c.ethClient.CallContract(ctx, ethereum.CallMsg{
			To:   &c.contractAddr,
			Data: callData,
		}, nil)
		if err == nil {
			break
		}

		time.Sleep(time.Duration((i+1)*200) * time.Millisecond)
	}
	if err != nil {
		return outputData, fmt.Errorf("failed to call contract, error: %v", err)
	}

	// 解析返回值
	if err = c.parsedABI.UnpackIntoInterface(&outputData, methodName, result); err != nil {
		return outputData, fmt.Errorf("failed to unpack result, error: %v", err)
	}

	return outputData, nil
}
