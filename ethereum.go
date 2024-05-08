package ethereumhelper

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/go-kratos/kratos/v2/log"
	"math/big"
)

type Converter[tx any] interface {
	ConvertTransaction(*types.Transaction) tx
}

type BlockFetcher[block any, tx any] interface {
	GetBlockNumber(ctx context.Context) (uint64, error)
	GetBlockHeaderByNumber(ctx context.Context, blockNumber uint64) (BlockHeader, error)
	GetBlockByNumber(ctx context.Context, targetBlock uint64) (Block[block, tx], error)
}

// 协议适配器
type ProtocolAdapter interface {
	String() string
	Validate() error
}

type Parser[tx any] interface {
	// 检查格式
	CheckFormat(data []byte) error
	// 解析
	Parse(tx) (ProtocolAdapter, error)
}

func New[block any, tx any](endpoints []string, parser Parser[tx], converter Converter[tx], logger log.Logger) (BlockFetcher[block, tx], error) {
	var clients []*ethclient.Client
	for _, endpoint := range endpoints {
		c, err := rpc.DialOptions(context.Background(), endpoint)
		if err != nil {
			return nil, err
		}

		clients = append(clients, ethclient.NewClient(c))
	}

	return &EthereumFetcher[block, tx]{
		clients:   clients,
		parser:    parser,
		logger:    log.NewHelper(log.With(logger, "module", "fetcher")),
		converter: converter,
	}, nil
}

// 区块头信息
type BlockHeader interface {
	GetNumber() uint64
	GetParentHash() string
	GetHash() string
}

type blockHeader struct {
	number     uint64 // 区块号
	parentHash string // 父区块哈希
	hash       string // 区块哈希
}

func (e *blockHeader) GetNumber() uint64 {
	return e.number
}

func (e *blockHeader) GetParentHash() string {
	return e.parentHash
}

func (e *blockHeader) GetHash() string {
	return e.hash
}

// 区块信息
type Block[block any, tx any] interface {
	GetHeader() BlockHeader
	GetTransactions() []Transaction[tx]
}

type BlockData[block any, tx any] struct {
	header       BlockHeader
	transactions []Transaction[tx]
}

func (b *BlockData[block, tx]) GetHeader() BlockHeader {
	return b.header

}
func (b *BlockData[block, tx]) GetTransactions() []Transaction[tx] {
	return b.transactions
}

func (b *BlockData[block, tx]) AppendTransaction(transaction Transaction[tx]) {
	b.transactions = append(b.transactions, transaction)
}

// 区块交易信息
type Transaction[tx any] interface {
	GetData() tx
	GetFrom() string
	GetTo() string
	GetPosition() string
}

type TransactionData[tx any] struct {
	data     tx
	from     string
	to       string
	position int
}

func (e *EthereumFetcher[block, tx]) NewTransaction(position int, transaction *types.Transaction) (Transaction[tx], error) {
	// 检查协议格式是否正确
	err := e.parser.CheckFormat(transaction.Data())
	if err != nil {
		return nil, err
	}

	var from common.Address
	if from, err = getTxSender(transaction); err != nil {
		return nil, err
	}

	var to common.Address
	if transaction.To() != nil {
		to = *transaction.To()
	}

	return &TransactionData[tx]{
		data:     e.converter.ConvertTransaction(transaction),
		from:     from.String(),
		to:       to.String(),
		position: position,
	}, nil
}

func (t *TransactionData[tx]) GetData() tx {
	return t.data
}
func (t *TransactionData[tx]) GetFrom() string {
	return t.from
}
func (t *TransactionData[tx]) GetTo() string {
	return t.to
}
func (t *TransactionData[tx]) GetPosition() string {
	return t.to
}

type EthereumFetcher[block any, tx any] struct {
	clients   []*ethclient.Client
	parser    Parser[tx]
	logger    *log.Helper
	converter Converter[tx]
}

func (e *EthereumFetcher[block, tx]) GetBlockNumber(ctx context.Context) (uint64, error) {
	return e.clients[0].BlockNumber(ctx)
}

func (e *EthereumFetcher[block, tx]) GetBlockHeaderByNumber(ctx context.Context, blockNumber uint64) (BlockHeader, error) {
	var params *big.Int
	if blockNumber != 0 {
		params = new(big.Int).SetUint64(blockNumber)
	}
	header, err := e.clients[0].HeaderByNumber(ctx, params)
	if err != nil {
		return nil, err
	}
	return &blockHeader{
		number:     header.Number.Uint64(),
		parentHash: header.ParentHash.String(),
		hash:       header.Hash().String(),
	}, nil
}

func (e *EthereumFetcher[block, tx]) GetBlockByNumber(ctx context.Context, targetBlock uint64) (Block[block, tx], error) {
	var currentBlock *types.Block

	for idx, cli := range e.clients {
		localClient := cli
		fetchedBlock, err := localClient.BlockByNumber(ctx, new(big.Int).SetUint64(targetBlock))
		if err != nil {
			return nil, err
		}

		if currentBlock != nil && (fetchedBlock.Hash() != currentBlock.Hash() || fetchedBlock.Transactions().Len() != currentBlock.Transactions().Len()) {
			e.logger.Warnf(
				"区块hash不一致. client_idx: %d, last_number: %d, last_hash: %v, tx_count: %d, current_number: %d, current_hash: %v, tx_count: %d",
				idx, currentBlock.NumberU64(), currentBlock.Hash(), currentBlock.Transactions().Len(), fetchedBlock.NumberU64(), fetchedBlock.Hash(), fetchedBlock.Transactions().Len(),
			)
			return nil, fmt.Errorf("block inconsistency detected at client index %d", idx)
		}

		currentBlock = fetchedBlock
	}

	b := &BlockData[block, tx]{}

	for position, transaction := range currentBlock.Transactions() {
		newTx, err := e.NewTransaction(position, transaction)
		if err != nil {
			return nil, err
		}
		b.AppendTransaction(newTx)
	}

	return b, nil
}

func getTxSender(tx *types.Transaction) (common.Address, error) {
	switch {
	case tx.Type() == types.AccessListTxType:
		return types.Sender(types.NewEIP2930Signer(tx.ChainId()), tx)
	case tx.Type() == types.DynamicFeeTxType:
		return types.Sender(types.NewLondonSigner(tx.ChainId()), tx)
	default:
		return types.Sender(types.NewEIP155Signer(tx.ChainId()), tx)
	}
}
