package blockchain

import (
	"amdzy/gochain/utils"
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

const targetBits = 15
const maxNonce = math.MaxInt64

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func (pow *ProofOfWork) PrepareDate(nonce int) ([]byte, error) {
	txHash, err := pow.block.HashTransactions()
	if err != nil {
		return nil, err
	}

	data := bytes.Join([][]byte{
		pow.block.PrevBlockHash,
		txHash,
		utils.IntToHex(pow.block.Timestamp),
		utils.IntToHex(targetBits),
		utils.IntToHex(int64(nonce)),
	}, []byte{})

	return data, nil
}

func (pow *ProofOfWork) Run() (int, []byte, error) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining a new block\n")

	for nonce < maxNonce {
		data, err := pow.PrepareDate(nonce)
		if err != nil {
			return 0, nil, err
		}
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash)

		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:], nil
}

func (pow *ProofOfWork) Validate() (bool, error) {
	var hashInt big.Int

	data, err := pow.PrepareDate(pow.block.Nonce)
	if err != nil {
		return false, err
	}
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1, nil
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}
