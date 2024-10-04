package blockchain

import (
	"amdzy/gochain/pkg/transactions"
)

type Blockchain struct {
	db  *DB
	tip []byte
}

func (bc *Blockchain) MineBlock(transactions []*transactions.Transaction) error {
	tip, err := bc.db.GetTip()
	if err != nil {
		return err
	}

	block := NewBlock(transactions, tip)
	err = bc.db.AddBlock(block)
	if err != nil {
		return err
	}

	bc.tip = block.Hash

	return nil
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

func (bc *Blockchain) CloseDB() {
	bc.db.Close()
}

func NewBlockchain() (*Blockchain, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, err
	}

	tip, err := db.GetTip()
	if err != nil {
		return nil, err
	}

	return &Blockchain{db, tip}, nil
}

func CreateBlockChain(address string) (*Blockchain, error) {
	db, err := InitDB(address)
	if err != nil {
		return nil, err
	}

	tip, err := db.GetTip()
	if err != nil {
		return nil, err
	}

	return &Blockchain{db, tip}, nil
}
