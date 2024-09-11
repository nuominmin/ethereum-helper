package ethereumhelper

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"io"
	"log"
	"math/big"
	"time"
)

type ContractReadHandler struct {
	ethClient    *ethclient.Client // 以太坊客户端
	parsedABI    abi.ABI           // 合约ABI
	contractAddr common.Address    // 合约地址
	retryNum     int               // 重试次数
}

func NewContractReadHandler(ethClient *ethclient.Client, contractAddr string, abiReader io.Reader) *ContractReadHandler {
	// 解析合约ABI
	parsedABI, err := abi.JSON(abiReader)
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	return &ContractReadHandler{
		ethClient:    ethClient,
		parsedABI:    parsedABI,
		contractAddr: common.HexToAddress(contractAddr),
	}
}

func (c *ContractReadHandler) SetRetryNum(n int) {
	c.retryNum = n + 1
}

func ContractRead[T any](ctx context.Context, c *ContractReadHandler, methodName string, args ...interface{}) (T, error) {
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

type ContractWriteHandler struct {
	*ContractReadHandler
	privateKeyECDSA *ecdsa.PrivateKey
	fromAddress     common.Address // 发起交易的地址
	chainId         *big.Int
}

func NewContractWriteHandler(ethClient *ethclient.Client, contractAddr string, abiReader io.Reader, privateKey string) *ContractWriteHandler {
	// 解析私钥
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		log.Fatalf("Invalid private key, error: %v", err)
	}

	// 获取当前网络的链ID
	chainId, err := ethClient.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("failed to get network ID, error: %v", err)
	}

	return &ContractWriteHandler{
		ContractReadHandler: NewContractReadHandler(ethClient, contractAddr, abiReader),
		privateKeyECDSA:     privateKeyECDSA,
		fromAddress:         crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		chainId:             chainId,
	}
}

func ContractWrite(ctx context.Context, c *ContractWriteHandler, methodName string, args ...interface{}) (common.Hash, error) {
	// 获取当前nonce
	nonce, err := c.ethClient.PendingNonceAt(ctx, c.fromAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get nonce, error: %v", err)
	}

	// 准备合约方法调用的数据
	var callData []byte
	if callData, err = c.parsedABI.Pack(methodName, args...); err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack method call, error: %v", err)
	}

	// 估算所需的 Gas 限制
	var gasLimit uint64
	gasLimit, err = c.ethClient.EstimateGas(ctx, ethereum.CallMsg{
		From: c.fromAddress,
		To:   &c.contractAddr,
		Data: callData,
	})
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to estimate gas limit: %v", err)
	}

	// 获取当前建议的 Gas 价格
	var gasPrice *big.Int
	if gasPrice, err = c.ethClient.SuggestGasPrice(ctx); err != nil {
		return common.Hash{}, fmt.Errorf("failed to get suggested gas price: %v", err)
	}

	// 创建交易对象
	tx := types.NewTransaction(nonce, c.contractAddr, big.NewInt(0), gasLimit, gasPrice, callData)

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainId), c.privateKeyECDSA)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction, error: %v", err)
	}

	// 发送交易
	err = c.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to send transaction, error: %v", err)
	}

	// 可选：等待交易确认
	var receipt *types.Receipt
	for {
		receipt, err = c.ethClient.TransactionReceipt(ctx, signedTx.Hash())
		if err == nil {
			return receipt.TxHash, nil
		}
		if errors.Is(err, ethereum.NotFound) {
			time.Sleep(time.Second * 2) // 等待2秒钟再查询
			continue
		}
		return common.Hash{}, fmt.Errorf("failed to get transaction receipt, error: %v", err)
	}
}
