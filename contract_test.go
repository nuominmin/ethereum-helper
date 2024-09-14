package ethereumhelper_test

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	ethereumhelper "github.com/nuominmin/ethereum-helper"
	"github.com/shopspring/decimal"
	"math/big"
	"testing"
)

func TestContractRead(t *testing.T) {
	rpcClient, err := rpc.Dial(ethereumhelper.EthRpc)
	if err != nil {
		t.Error(err)
		return
	}
	ethClient := ethclient.NewClient(rpcClient)
	contract := ethereumhelper.NewContractHandler(ethClient)
	abiJson := ethereumhelper.ContractAbiJson
	contractAddr := common.HexToAddress(ethereumhelper.ContractAddr)
	methodName := ethereumhelper.ContractReadMethodName

	var res *big.Int
	if res, err = ethereumhelper.ContractRead[*big.Int](context.Background(), contract, abiJson, contractAddr, methodName); err != nil {
		t.Error(err)
		return
	}

	t.Log(ethereumhelper.ContractReadMethodName, decimal.NewFromBigInt(res, 0).Div(decimal.NewFromInt(1000)))
}

func TestContractWrite(t *testing.T) {
	rpcClient, err := rpc.Dial(ethereumhelper.EthRpc)
	if err != nil {
		t.Error(err)
		return
	}
	ethClient := ethclient.NewClient(rpcClient)
	contract := ethereumhelper.NewContractHandler(ethClient)
	abi := ethereumhelper.ContractAbiJson
	contractAddr := common.HexToAddress(ethereumhelper.ContractAddr)
	methodName := ethereumhelper.ContractWriteMethodName
	privateKey := ethereumhelper.PrivateKey

	args := []interface{}{common.HexToAddress(ethereumhelper.ContractAddr), big.NewInt(2)}

	var txHash common.Hash
	txHash, err = ethereumhelper.ContractWrite(context.Background(), contract, abi, contractAddr, privateKey, methodName, args...)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%s, https://sepolia.etherscan.io/tx/%s", ethereumhelper.ContractWriteMethodName, txHash.String())
}
