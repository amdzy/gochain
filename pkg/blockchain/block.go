package blockchain

import (
	"amdzy/gochain/pkg/transactions"
	"bytes"
	"crypto/sha256"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type Block struct {
	Timestamp     int64
	Transactions  []*transactions.Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func (block *Block) Serialize() ([]byte, error) {
	b, err := msgpack.Marshal(block)

	return b, err
}

func (block *Block) HashTransactions() []byte {
	var hashes [][]byte

	for _, tx := range block.Transactions {
		hashes = append(hashes, tx.ID)
	}

	hash := sha256.Sum256(bytes.Join(hashes, []byte{}))

	return hash[:]
}

func DeserializeBlock(b []byte) (*Block, error) {
	var block Block
	err := msgpack.Unmarshal(b, &block)

	return &block, err
}

func NewBlock(transactions []*transactions.Transaction, prevHash []byte) *Block {
	block := &Block{
		Transactions:  transactions,
		PrevBlockHash: prevHash,
		Timestamp:     time.Now().UTC().Unix(),
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func NewGenesisBlock(coinbase *transactions.Transaction) *Block {
	return NewBlock([]*transactions.Transaction{coinbase}, []byte{})
}
