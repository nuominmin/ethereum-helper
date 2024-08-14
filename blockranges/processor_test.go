package blockranges

import (
	"fmt"
	"testing"
)

func TestBlockRangeProcessor_Handle(t *testing.T) {
	// 2000 A
	// 3000 B
	// 4000 C

	// f(2500) B

	type Conf struct {
		WalletAddr   string
		ContractAddr string
		CodeVersion  string
	}

	ranges := []Range[Conf]{}
	ranges = append(ranges, Range[Conf]{StartBlock: 3000, Data: Conf{
		WalletAddr: "BBBB", ContractAddr: "BBBB", CodeVersion: "v2",
	}})
	ranges = append(ranges, Range[Conf]{StartBlock: 2000, Data: Conf{
		WalletAddr: "AAAA", ContractAddr: "AAAA", CodeVersion: "v1",
	}})
	ranges = append(ranges, Range[Conf]{StartBlock: 4500, Data: Conf{
		WalletAddr: "CCCC", ContractAddr: "CCCC", CodeVersion: "v4",
	}})
	ranges = append(ranges, Range[Conf]{StartBlock: 4000, Data: Conf{
		WalletAddr: "CCCC", ContractAddr: "CCCC", CodeVersion: "v3",
	}})

	p, err := NewBlockRangeProcessor[Conf](ranges...)
	if err != nil {
		t.Errorf("NewBlockRangeProcessor[Conf] error: %s", err)
		return
	}

	_ = p.Handle(1000, func(data Conf) error {
		fmt.Println(data)
		return nil
	})

	_ = p.Handle(2500, func(data Conf) error {
		fmt.Println(data)
		return nil
	})

	_ = p.Handle(3000, func(data Conf) error {
		fmt.Println(data)
		return nil
	})

	_ = p.Handle(5000, func(data Conf) error {
		fmt.Println(data)
		return nil
	})

}
