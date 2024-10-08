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

func (block *Block) HashTransactions() ([]byte, error) {
	var hashes [][]byte

	for _, tx := range block.Transactions {
		b, err := tx.Hash()
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, b)
	}

	hash := sha256.Sum256(bytes.Join(hashes, []byte{}))

	return hash[:], nil
}

func DeserializeBlock(b []byte) (*Block, error) {
	var block Block
	err := msgpack.Unmarshal(b, &block)

	return &block, err
}

func NewBlock(transactions []*transactions.Transaction, prevHash []byte) (*Block, error) {
	block := &Block{
		Transactions:  transactions,
		PrevBlockHash: prevHash,
		Timestamp:     time.Now().UTC().Unix(),
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
	return NewBlock([]*transactions.Transaction{coinbase}, []byte{})
}
