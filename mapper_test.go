package ethereumhelper

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"
	"testing"
)

func TestAddress(t *testing.T) {
	a := new(big.Int)
	a.SetString("65625641663004913611543007257256861518297599113210899750616375382861399306046", 10)

	hash := fmt.Sprintf("0x%s", hex.EncodeToString(a.Bytes()))
	fmt.Println(hash, common.BigToHash(a).String()) // 0x9116cc00fe051805c1433e9705898afd22519cac1a5222d022a9a08530d9af3e

	b := common.HexToHash("0x9116cc00fe051805c1433e9705898afd22519cac1a5222d022a9a08530d9af3e").Big()
	fmt.Println(b, a.String() == b.String())
}

func TestHex(t *testing.T) {
	a := "0x65746869"
	bytesTick, err := hex.DecodeString(strings.TrimLeft(a, "0x"))
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(string(bytesTick)) // ethi

	data := "ethi"
	b := fmt.Sprintf("0x%s", hex.EncodeToString([]byte(data)))
	fmt.Println(b, a == b)
}
