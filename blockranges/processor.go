package blockranges

import (
	"errors"
	"sort"
)

// Range 范围
type Range[T any] struct {
	StartBlock uint64 // 包含
	Data       T
}

// ErrOverlappingRanges 用于表示区间重叠错误
var ErrOverlappingRanges = errors.New("config have overlapping ranges")

// ErrInvalidStartBlock 用于表示起始块无效错误
var ErrInvalidStartBlock = errors.New("start block must be greater than 0")

// ErrRangeIsEmpty range 是空的
var ErrRangeIsEmpty = errors.New("range is empty")

type BlockRangeProcessor[T any] struct {
	ranges []Range[T]
}

func NewBlockRangeProcessor[T any]() *BlockRangeProcessor[T] {
	return &BlockRangeProcessor[T]{}
}

// AddRange 添加一个新的范围到 BlockRangeProcessor 中
func (pb *BlockRangeProcessor[T]) AddRange(r Range[T]) error {
	if r.StartBlock == 0 {
		return ErrInvalidStartBlock
	}

	sort.Slice(pb.ranges, func(i, j int) bool {
		// 根据 StartBlock 降序
		return pb.ranges[i].StartBlock > pb.ranges[j].StartBlock
	})

	for i := 0; i < len(pb.ranges); i++ {
		if pb.ranges[i].StartBlock == 0 {
			return ErrInvalidStartBlock
		}
		if i > 0 && pb.ranges[i].StartBlock == pb.ranges[i-1].StartBlock {
			return ErrOverlappingRanges
		}
	}

	return nil
}

func (pb *BlockRangeProcessor[T]) Handle(blockNumber uint64, handler func(data T) error) error {
	for i := 0; i < len(pb.ranges); i++ {
		if blockNumber >= pb.ranges[i].StartBlock {
			return handler(pb.ranges[i].Data)
		}
	}

	var data T
	return handler(data)
}
