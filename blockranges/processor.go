package blockranges

import (
	"errors"
	"sort"
)

// Range 范围
type Range[T any] struct {
	StartBlock uint64 // 包含
	data       T
}

// ErrOverlappingRanges 用于表示区间重叠错误
var ErrOverlappingRanges = errors.New("config have overlapping ranges")

// ErrInvalidStartBlock 用于表示起始块无效错误
var ErrInvalidStartBlock = errors.New("start block must be greater than 0")

type BlockRangeProcessor[T any] struct {
	ranges []Range[T]
}

func NewBlockRangeProcessor[T any](ranges ...Range[T]) (*BlockRangeProcessor[T], error) {
	sort.Slice(ranges, func(i, j int) bool {
		// 根据 StartBlock 降序
		return ranges[i].StartBlock > ranges[j].StartBlock
	})

	for i := 0; i < len(ranges); i++ {
		if ranges[i].StartBlock == 0 {
			return nil, ErrInvalidStartBlock
		}
		if i > 0 && ranges[i].StartBlock == ranges[i-1].StartBlock {
			return nil, ErrOverlappingRanges
		}
	}

	return &BlockRangeProcessor[T]{
		ranges: ranges,
	}, nil
}

func (pb *BlockRangeProcessor[T]) Handle(blockNumber uint64, handler func(data T) error) error {
	for i := 0; i < len(pb.ranges); i++ {
		if blockNumber >= pb.ranges[i].StartBlock {
			return handler(pb.ranges[i].data)
		}
	}

	var data T
	return handler(data)
}
