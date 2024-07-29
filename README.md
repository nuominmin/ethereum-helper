### 示例

``` go
package ierc

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethereumhelper "github.com/nuominmin/ethereum-helper"
	"math/big"
)

type TicketUsed struct {
	From   common.Address
	Tick   [4]byte
	Amount *big.Int
	TxId   *big.Int
}

func UnpackTicketUsed(parsedABI abi.ABI, log *types.Log) (event *TicketUsed, err error) {
	return ethereumhelper.Unpack[TicketUsed](parsedABI, "TicketUsed", log)
}

type Swap struct {
	To         common.Address
	Amount0In  *big.Int
	Amount1In  *big.Int
	Amount0Out *big.Int
	Amount1Out *big.Int
}

func UnpackSwap(parsedABI abi.ABI, log *types.Log) (event *Swap, err error) {
	return ethereumhelper.Unpack[Swap](parsedABI, "Swap", log)
}

type TickTransfer struct {
	To     common.Address
	Tick   [4]byte
	Amount *big.Int
}

func UnpackTickTransfer(parsedABI abi.ABI, log *types.Log) (event *TickTransfer, err error) {
	return ethereumhelper.Unpack[TickTransfer](parsedABI, "TickTransfer", log)
}

type Mint struct {
	To      common.Address
	Amount0 *big.Int
	Amount1 *big.Int
}

func UnpackMint(parsedABI abi.ABI, log *types.Log) (event *Mint, err error) {
	return ethereumhelper.Unpack[Mint](parsedABI, "Mint", log)
}

type Burn struct {
	To      common.Address
	Amount0 *big.Int
	Amount1 *big.Int
}

func UnpackBurn(parsedABI abi.ABI, log *types.Log) (event *Burn, err error) {
	return ethereumhelper.Unpack[Burn](parsedABI, "Burn", log)
}
```
