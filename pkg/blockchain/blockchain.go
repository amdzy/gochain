package blockchain

import (
	"amdzy/gochain/pkg/transactions"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
)

type Blockchain struct {
	Db  *DB
	tip []byte
}

func (bc *Blockchain) MineBlock(transactions []*transactions.Transaction) (*Block, error) {
	tip, err := bc.Db.GetTip()
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

	block, err := NewBlock(transactions, tip)
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

func (bc *Blockchain) CloseDB() {
	bc.Db.Close()
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
