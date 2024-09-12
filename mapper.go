package ethereumhelper

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
)

// TxIdToHash 65625641663004913611543007257256861518297599113210899750616375382861399306046 => 0x9116cc00fe051805c1433e9705898afd22519cac1a5222d022a9a08530d9af3e
func TxIdToHash(txId *big.Int) string {
	return common.BigToHash(txId).String()
}

// TxHashToId 0x9116cc00fe051805c1433e9705898afd22519cac1a5222d022a9a08530d9af3e => 65625641663004913611543007257256861518297599113210899750616375382861399306046
func TxHashToId(txHash string) *big.Int {
	return common.HexToHash("0x9116cc00fe051805c1433e9705898afd22519cac1a5222d022a9a08530d9af3e").Big()
}

// HexToStr 0x65746869 => ethi
func HexToStr(s string) (string, error) {
	bytes, err := hex.DecodeString(strings.TrimLeft(s, "0x"))
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// StrToHex ethi => 0x65746869
func StrToHex(s string) string {
	return fmt.Sprintf("0x%s", hex.EncodeToString([]byte(s)))
}
