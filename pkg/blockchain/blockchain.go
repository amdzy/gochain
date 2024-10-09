package blockchain

import (
	"amdzy/gochain/pkg/transactions"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type Blockchain struct {
	Db  *DB
	tip []byte
}

func (bc *Blockchain) MineBlock(transactions []*transactions.Transaction) (*Block, error) {
	lastHash, lastHeight, err := bc.Db.GetLastHashAndHeight()
	if err != nil {
		return nil, err
	}

	for _, tx := range transactions {
		verified, err := bc.VerifyTransaction(tx)
		if err != nil {
			return nil, err
		}

		if !verified {
			return nil, fmt.Errorf("invalid transaction")
		}
	}

	block, err := NewBlock(transactions, lastHash, lastHeight+1)
	if err != nil {
		return nil, err
	}

	err = bc.Db.AddBlock(block)
	if err != nil {
		return nil, err
	}

	bc.tip = block.Hash

	return block, nil
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.Db}

	return bci
}

func (bc *Blockchain) FindUTXO() (map[string]transactions.TXOutputs, error) {
	UTXO := make(map[string]transactions.TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			return nil, err
		}

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return UTXO, nil
}

func (bc *Blockchain) FindTransaction(id []byte) (*transactions.Transaction, error) {
	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			return nil, err
		}

		for _, tx := range block.Transactions {
			if bytes.Equal(id, tx.ID) {
				return tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return nil, fmt.Errorf("transaction not found")
}

func (bc *Blockchain) SignTransaction(tx *transactions.Transaction, privKey ecdsa.PrivateKey) error {
	prevTXs := make(map[string]transactions.Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			return err
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = *prevTX
	}

	return tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx *transactions.Transaction) (bool, error) {
	if tx.IsCoinbase() {
		return true, nil
	}

	prevTXs := make(map[string]transactions.Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		if err != nil {
			return false, nil
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = *prevTX
	}

	return tx.Verify(prevTXs)
}

func (bc *Blockchain) AddBlock(block *Block) error {
	err := bc.Db.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData, err := block.Serialize()
		if err != nil {
			return err
		}

		err = b.Put(block.Hash, blockData)
		if err != nil {
			return err
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock, err := DeserializeBlock(lastBlockData)
		if err != nil {
			return err
		}

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				return err
			}
			bc.tip = block.Hash
		}

		return nil
	})

	return err
}

func (bc *Blockchain) GetBestHeight() (int, error) {
	var lastBlock *Block

	err := bc.Db.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		var err error
		lastBlock, err = DeserializeBlock(blockData)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return lastBlock.Height, nil
}

func (bc *Blockchain) GetBlock(blockHash []byte) (*Block, error) {
	var block *Block

	err := bc.Db.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("block not found")
		}

		var err error
		block, err = DeserializeBlock(blockData)
		if err != nil {
			return errors.New("block not found")
		}

		return nil
	})

	return block, err
}

func (bc *Blockchain) GetBlockHashes() ([][]byte, error) {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks, nil
}

func (bc *Blockchain) CloseDB() {
	bc.Db.Close()
}

func NewBlockchain() (*Blockchain, error) {
	db, err := ConnectDB()
	if err != nil {
		return nil, err
	}

	lastHash, _, err := db.GetLastHashAndHeight()
	if err != nil {
		return nil, err
	}

	return &Blockchain{db, lastHash}, nil
}

func CreateBlockChain(address string) (*Blockchain, error) {
	db, err := InitDB(address)
	if err != nil {
		return nil, err
	}

	lastHash, _, err := db.GetLastHashAndHeight()
	if err != nil {
		return nil, err
	}

	return &Blockchain{db, lastHash}, nil
}
