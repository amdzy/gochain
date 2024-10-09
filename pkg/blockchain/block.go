package blockchain

import (
	"amdzy/gochain/pkg/merkle"
	"amdzy/gochain/pkg/transactions"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type Block struct {
	Timestamp     int64
	Transactions  []*transactions.Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height        int
}

func (block *Block) Serialize() ([]byte, error) {
	b, err := msgpack.Marshal(block)

	return b, err
}

func (block *Block) HashTransactions() ([]byte, error) {
	var transactions [][]byte

	for _, tx := range block.Transactions {
		txSerialized, err := tx.Serialize()
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, txSerialized)
	}
	mTree := merkle.NewMerkleTree(transactions)

	return mTree.RootNode.Data, nil
}

func DeserializeBlock(b []byte) (*Block, error) {
	var block Block
	err := msgpack.Unmarshal(b, &block)

	return &block, err
}

func NewBlock(transactions []*transactions.Transaction, prevHash []byte, height int) (*Block, error) {
	block := &Block{
		Transactions:  transactions,
		PrevBlockHash: prevHash,
		Timestamp:     time.Now().UTC().Unix(),
		Height:        height,
	}

	pow := NewProofOfWork(block)
	nonce, hash, err := pow.Run()
	if err != nil {
		return nil, err
	}

	block.Hash = hash[:]
	block.Nonce = nonce

	return block, nil
}

func NewGenesisBlock(coinbase *transactions.Transaction) (*Block, error) {
	return NewBlock([]*transactions.Transaction{coinbase}, []byte{}, 0)
}
