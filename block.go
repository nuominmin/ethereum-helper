package ethereumhelper

import (
	"github.com/ethereum/go-ethereum/core/types"
	"time"
)

func GetBlockNumber(block *types.Block) uint64 {
	if block == nil {
		return 0
	}
	return block.NumberU64()
}

func GetBlockTimeUnix(block *types.Block) int64 {
	if block == nil {
		return 0
	}
	return int64(block.Time())
}

func GetBlockTime(block *types.Block) time.Time {
	if block == nil {
		return time.Time{}
	}
	return time.Unix(int64(block.Time()), 0)
}
