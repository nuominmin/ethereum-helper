package ethereumhelper_test

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	ethereumhelper "github.com/nuominmin/ethereum-helper"
	"github.com/shopspring/decimal"
	"math/big"
	"strings"
	"testing"
)

func TestContractRead(t *testing.T) {
	rpcClient, err := rpc.Dial(ethereumhelper.EthRpc)
	if err != nil {
		t.Error(err)
		return
	}
	ethClient := ethclient.NewClient(rpcClient)
	contract := ethereumhelper.NewContractReadHandler(ethClient, ethereumhelper.ContractAddr, strings.NewReader(ethereumhelper.ContractAbiJson))

	var res *big.Int
	if res, err = ethereumhelper.ContractRead[*big.Int](context.Background(), contract, ethereumhelper.ContractReadMethodName); err != nil {
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
	contract := ethereumhelper.NewContractWriteHandler(ethClient, ethereumhelper.ContractAddr, strings.NewReader(ethereumhelper.ContractAbiJson), ethereumhelper.PrivateKey)

	var txHash common.Hash
	txHash, err = ethereumhelper.ContractWrite(context.Background(), contract, ethereumhelper.ContractWriteMethodName, common.HexToAddress(ethereumhelper.ContractAddr), big.NewInt(2))
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(ethereumhelper.ContractWriteMethodName, txHash.String())
}
