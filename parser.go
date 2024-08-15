package ethereumhelper

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"io"
	"log"
)

type Transactions []Transaction
type Transaction interface {
	SetPositionInTxs(i int64)
	GetTxHash() string
	GetBlockNumber() uint64
	GetTime() int64
}

func TransactionsLen[T any](txs Transactions) int {
	var n int
	for i := 0; i < len(txs); i++ {
		if _, ok := txs[i].(T); ok {
			n++
		}
	}
	return n
}

// 交易方法解析
type TxMethodParser interface {
	MethodName() string
	Parse(parsedABI abi.ABI, block *types.Block, transaction *types.Transaction, txLogs []*types.Log) (Transactions, error)
}

// 合约事件解析
type ContractEventParser interface {
	Parse(parsedABI abi.ABI, block *types.Block, txLog *types.Log) (Transaction, error)
}

type Parser struct {
	parsedABI     abi.ABI
	methodParsers []TxMethodParser
	eventParsers  []ContractEventParser
}

func NewParser(abiReader io.Reader) *Parser {
	// 解析合约ABI
	parsedABI, err := abi.JSON(abiReader)
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	return &Parser{
		parsedABI: parsedABI,
	}
}

func (p *Parser) AddTxMethodParsers(parsers ...TxMethodParser) *Parser {
	p.methodParsers = append(p.methodParsers, parsers...)
	return p
}

func (p *Parser) AddContractEventParser(parsers ...ContractEventParser) *Parser {
	p.eventParsers = append(p.eventParsers, parsers...)
	return p
}

func (p *Parser) ParserMethods(block *types.Block, transaction *types.Transaction, txLogs []*types.Log) Transactions {
	method, err := p.parsedABI.MethodById(transaction.Data())
	if err != nil {
		return nil
	}

	var transactions Transactions
	for i := 0; i < len(p.methodParsers); i++ {
		if method.Name != p.methodParsers[i].MethodName() {
			continue
		}

		var txs Transactions
		if txs, err = p.methodParsers[i].Parse(p.parsedABI, block, transaction, txLogs); err != nil {
			continue
		}
		if len(txs) > 0 {
			transactions = append(transactions, txs...)
		}
	}
	return transactions
}

func (p *Parser) ParserEvents(block *types.Block, txLogs []*types.Log) Transactions {
	var events Transactions
	for i := 0; i < len(txLogs); i++ {
		for j := 0; j < len(p.eventParsers); j++ {
			event, err := p.eventParsers[j].Parse(p.parsedABI, block, txLogs[i])
			if err != nil {
				continue
			}
			if event != nil {
				events = append(events, event)
			}
		}
	}
	return events
}

// Close 释放资源
func (p *Parser) Close() error {
	p.methodParsers = nil
	p.eventParsers = nil
	return nil
}
