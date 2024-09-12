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

type ContractHandler struct {
	ethClient *ethclient.Client // 以太坊客户端
	retryNum  int               // 重试次数
	chainId   *big.Int
}

func NewContractHandler(ethClient *ethclient.Client) *ContractHandler {
	// 获取当前网络的链ID
	chainId, err := ethClient.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("failed to get network ID, error: %v", err)
	}

	return &ContractHandler{
		ethClient: ethClient,
		retryNum:  1,
		chainId:   chainId,
	}
}

func (c *ContractHandler) SetRetryNum(n int) {
	c.retryNum = n + 1
}

func ContractRead[T any](ctx context.Context, ch *ContractHandler, abiReader io.Reader, contractAddr common.Address, methodName string, args ...interface{}) (T, error) {
	var outputData T

	// 解析合约ABI
	parsedABI, err := abi.JSON(abiReader)
	if err != nil {
		return outputData, fmt.Errorf("failed to parse contract ABI, error: %v", err)
	}

	// 准备方法调用
	var callData []byte
	if callData, err = parsedABI.Pack(methodName, args...); err != nil {
		return outputData, fmt.Errorf("failed to parse abi, error: %v", err)
	}

	// 执行调用
	var result []byte
	for i := 0; i < ch.retryNum; i++ {
		result, err = ch.ethClient.CallContract(ctx, ethereum.CallMsg{
			To:   &contractAddr,
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
	if err = parsedABI.UnpackIntoInterface(&outputData, methodName, result); err != nil {
		return outputData, fmt.Errorf("failed to unpack result, error: %v", err)
	}

	return outputData, nil
}

func ContractWrite(ctx context.Context, ch *ContractHandler, abiReader io.Reader, contractAddr common.Address, privateKey string, methodName string, args ...interface{}) (common.Hash, error) {
	// 解析合约ABI
	parsedABI, err := abi.JSON(abiReader)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to parse contract ABI, error: %v", err)
	}

	// 解析私钥
	var privateKeyECDSA *ecdsa.PrivateKey
	if privateKeyECDSA, err = crypto.HexToECDSA(privateKey); err != nil {
		return common.Hash{}, fmt.Errorf("invalid private key, error: %v", err)
	}

	// 发起交易的地址
	fromAddress := crypto.PubkeyToAddress(privateKeyECDSA.PublicKey)

	// 获取当前nonce
	var nonce uint64
	if nonce, err = ch.ethClient.PendingNonceAt(ctx, fromAddress); err != nil {
		return common.Hash{}, fmt.Errorf("failed to get nonce, error: %v", err)
	}

	// 准备合约方法调用的数据
	var callData []byte
	if callData, err = parsedABI.Pack(methodName, args...); err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack method call, error: %v", err)
	}

	// 估算所需的 Gas 限制
	var gasLimit uint64
	gasLimit, err = ch.ethClient.EstimateGas(ctx, ethereum.CallMsg{
		From: fromAddress,
		To:   &contractAddr,
		Data: callData,
	})
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to estimate gas limit, error: %v", err)
	}

	// 获取当前建议的 Gas 价格
	var gasPrice *big.Int
	if gasPrice, err = ch.ethClient.SuggestGasPrice(ctx); err != nil {
		return common.Hash{}, fmt.Errorf("failed to get suggested gas price, error: %v", err)
	}

	// 创建交易对象
	tx := types.NewTransaction(nonce, contractAddr, big.NewInt(0), gasLimit, gasPrice, callData)

	// 签名交易
	var signedTx *types.Transaction
	if signedTx, err = types.SignTx(tx, types.NewEIP155Signer(ch.chainId), privateKeyECDSA); err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction, error: %v", err)
	}

	// 发送交易
	if err = ch.ethClient.SendTransaction(ctx, signedTx); err != nil {
		return common.Hash{}, fmt.Errorf("failed to send transaction, error: %v", err)
	}

	// 可选：等待交易确认
	var receipt *types.Receipt
	for {
		if receipt, err = ch.ethClient.TransactionReceipt(ctx, signedTx.Hash()); err == nil {
			return receipt.TxHash, nil
		}
		if errors.Is(err, ethereum.NotFound) {
			time.Sleep(time.Second * 2) // 等待2秒钟再查询
			continue
		}
		return common.Hash{}, fmt.Errorf("failed to get transaction receipt, error: %v", err)
	}
}
