package blockchain

import (
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func (block *Block) Serialize() ([]byte, error) {
	b, err := msgpack.Marshal(block)

	return b, err
}

func DeserializeBlock(b []byte) (*Block, error) {
	var block Block
	err := msgpack.Unmarshal(b, &block)

	return &block, err
}

func NewBlock(data string, prevHash []byte) *Block {
	block := &Block{
		Data:          []byte(data),
		PrevBlockHash: prevHash,
		Timestamp:     time.Now().UTC().Unix(),
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
