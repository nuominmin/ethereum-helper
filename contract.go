package ethereumhelper

import (
	"bytes"
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
	"log"
	"math/big"
	"strings"
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
		retryNum:  3,
		chainId:   chainId,
	}
}

func (c *ContractHandler) SetRetryNum(n int) {
	c.retryNum = n + 1
}

func ContractRead[T any](ctx context.Context, ch *ContractHandler, abiJson string, contractAddr common.Address, methodName string, args ...interface{}) (T, error) {
	var outputData T

	// 解析合约ABI
	parsedABI, err := abi.JSON(strings.NewReader(abiJson))
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

func ContractWrite(ctx context.Context, ch *ContractHandler, abiJson string, contractAddr common.Address, privateKey string, methodName string, args ...interface{}) (common.Hash, error) {
	// 解析合约ABI
	parsedABI, err := abi.JSON(strings.NewReader(abiJson))
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

	// 获取当前区块的 Gas 限制
	var gasLimit uint64
	if gasLimit, err = getGasLimit(ctx, ch, fromAddress, contractAddr, callData); err != nil {
		return common.Hash{}, fmt.Errorf("failed to get gas limit, error: %v", err)
	}

	// 获取当前建议的 Gas 价格
	var gasPrice *big.Int
	if gasPrice, err = getGasPrice(ctx, ch); err != nil {
		return common.Hash{}, fmt.Errorf("failed to get suggested gas price, error: %v", err)
	}

	// 创建交易对象
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &contractAddr,
		Value:    big.NewInt(0),
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     callData,
	})

	// 签名交易
	var signedTx *types.Transaction
	if signedTx, err = types.SignTx(tx, types.NewEIP155Signer(ch.chainId), privateKeyECDSA); err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction, error: %v", err)
	}

	for i := 0; i < ch.retryNum; i++ {
		if err = ch.ethClient.SendTransaction(ctx, signedTx); err == nil {
			break
		}

		if strings.Contains(err.Error(), "transaction underpriced") {
			gasPrice = gasPrice.Mul(gasPrice, big.NewInt(2)) // 每次失败时将 gas price 提高一倍
			tx = types.NewTx(&types.LegacyTx{
				Nonce:    nonce,
				To:       &contractAddr,
				Value:    big.NewInt(0),
				Gas:      gasLimit,
				GasPrice: gasPrice,
				Data:     callData,
			})
			signedTx, err = types.SignTx(tx, types.NewEIP155Signer(ch.chainId), privateKeyECDSA)
			if err != nil {
				return common.Hash{}, fmt.Errorf("failed to sign transaction after gas adjustment, error: %v", err)
			}
		} else {
			return common.Hash{}, fmt.Errorf("failed to send transaction, error: %v", err)
		}

		time.Sleep(time.Duration((i+1)*200) * time.Millisecond)
	}

	fmt.Println("signedTx.Hash()", signedTx.Hash())
	var receipt *types.Receipt

	timeout := 30 * time.Second
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	for {
		select {
		case <-ctxWithTimeout.Done():
			return common.Hash{}, fmt.Errorf("transaction receipt retrieval timed out after %v", timeout)
		default:
			if receipt, err = ch.ethClient.TransactionReceipt(ctx, signedTx.Hash()); err == nil {
				if receipt.Status == types.ReceiptStatusFailed {

					msg := ethereum.CallMsg{
						From:     fromAddress,   // 从哪个地址发起调用
						To:       &contractAddr, // 合约地址
						Gas:      gasLimit,      // 让客户端估算所需的Gas
						GasPrice: gasPrice,      // Gas价格设置为0，因为只是模拟执行
						Value:    big.NewInt(0), // 发送的价值为0
						Data:     callData,      // 调用合约的数据
					}

					var res []byte
					if res, err = ch.ethClient.CallContract(ctx, msg, nil); err != nil {
						return common.Hash{}, fmt.Errorf("failed to retrieve revert reason, error: %v", err)
					}

					return common.Hash{}, fmt.Errorf("contract execution failed. %s", string(res))
				}

				return receipt.TxHash, nil
			}

			if errors.Is(err, ethereum.NotFound) {
				time.Sleep(time.Second * 1)
				continue
			}
			return common.Hash{}, fmt.Errorf("failed to get transaction receipt, error: %v", err)
		}
	}
}

func getGasLimit(ctx context.Context, ch *ContractHandler, fromAddress, contractAddr common.Address, callData []byte) (uint64, error) {
	// 估算所需的 Gas 限制
	gasLimit, err := ch.ethClient.EstimateGas(ctx, ethereum.CallMsg{
		From: fromAddress,
		To:   &contractAddr,
		Data: callData,
	})
	if err == nil {
		return gasLimit, nil
	}

	var header *types.Header
	if header, err = ch.ethClient.HeaderByNumber(ctx, nil); err != nil {
		return 0, fmt.Errorf("failed to get header, error: %v", err)
	}

	//gasLimit := header.GasLimit - 50000 // 减少 50000 个 Gas 单位，避免超过区块的 Gas 限制，导致交易失败

	return header.GasLimit, nil
}

func getGasPrice(ctx context.Context, ch *ContractHandler) (*big.Int, error) {
	// 获取当前建议的 Gas 价格
	gasPrice, err := ch.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	//gasPrice = gasPrice.Mul(gasPrice, big.NewInt(2))
	return gasPrice, nil
}

// ErrorSignature 是标准的 Error(string) 函数的签名哈希
var ErrorSignature = []byte{0x08, 0xc3, 0x79, 0xa0} // keccak256("Error(string)") 的前 4 个字节

func parseRevertReason(result []byte) (string, error) {
	if len(result) < 4 || !bytes.Equal(result[:4], ErrorSignature) {
		return "", fmt.Errorf("no revert reason")
	}

	// 使用 abi.NewType 来创建 string 类型
	stringType, err := abi.NewType("string", "", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create string ABI type: %v", err)
	}

	// 剩余部分是ABI编码的错误消息，尝试解码为字符串
	errorAbi, err := abi.Arguments{
		{Type: stringType},
	}.Unpack(result[4:])
	if err != nil {
		return "", fmt.Errorf("failed to unpack revert reason: %v", err)
	}

	if len(errorAbi) > 0 {
		if reason, ok := errorAbi[0].(string); ok {
			return reason, nil
		}
	}

	return "", fmt.Errorf("failed to parse revert reason")
}
