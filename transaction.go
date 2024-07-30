package ethereumhelper

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/shopspring/decimal"
	"math/big"
)

func GetTxHash(tx *types.Transaction) string {
	if tx == nil {
		return ""
	}

	return tx.Hash().String()
}

func GetTxValue(tx *types.Transaction, exp int32) decimal.Decimal {
	if tx == nil {
		return decimal.Zero
	}

	return decimal.NewFromBigInt(tx.Value(), exp)
}

func GetTxGasPrice(tx *types.Transaction, exp int32) decimal.Decimal {
	if tx == nil {
		return decimal.Zero
	}

	return decimal.NewFromBigInt(tx.GasPrice(), exp)
}

func GetTxGas(tx *types.Transaction, exp int32) decimal.Decimal {
	if tx == nil {
		return decimal.Zero
	}

	return decimal.NewFromBigInt(new(big.Int).SetUint64(tx.Gas()), 0)
}

func GetTxNonce(tx *types.Transaction) uint64 {
	if tx == nil {
		return 0
	}

	return tx.Nonce()
}

func GetTxSender(tx *types.Transaction) (common.Address, error) {
	if tx == nil {
		return common.Address{}, nil
	}

	var signer types.Signer
	switch tx.Type() {
	case types.LegacyTxType:
		signer = types.NewEIP155Signer(tx.ChainId())
	case types.AccessListTxType:
		signer = types.NewEIP2930Signer(tx.ChainId())
	case types.DynamicFeeTxType:
		signer = types.NewLondonSigner(tx.ChainId())
	default:
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	sender, err := types.Sender(signer, tx)
	return sender, err
}
