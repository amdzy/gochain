package utxo

import (
	"amdzy/gochain/pkg/blockchain"
	"amdzy/gochain/pkg/transactions"
	"amdzy/gochain/pkg/wallet"
	"encoding/hex"
	"fmt"
	"log"

	bolt "go.etcd.io/bbolt"
)

const utxoBucket = "utxo"

type UTXOSet struct {
	Blockchain *blockchain.Blockchain
}

func (u UTXOSet) ReIndex() error {
	db := u.Blockchain.Db.Db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			return err
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	UTXO, err := u.Blockchain.FindUTXO()
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}

			outsSerialized, err := outs.Serialize()
			if err != nil {
				return err
			}

			err = b.Put(key, outsSerialized)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int, error) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.Db.Db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs, err := transactions.DeserializeOutputs(v)
			if err != nil {
				return err
			}

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})

	if err != nil {
		return 0, nil, err
	}

	return accumulated, unspentOutputs, nil
}

func (u UTXOSet) FindUTXO(pubKeyHash []byte) ([]transactions.TXOutput, error) {
	var UTXOs []transactions.TXOutput
	db := u.Blockchain.Db.Db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs, err := transactions.DeserializeOutputs(v)
			if err != nil {
				return err
			}

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return UTXOs, nil
}

func (u UTXOSet) CountTransactions() (int, error) {
	db := u.Blockchain.Db.Db
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return counter, nil
}

func (u UTXOSet) Update(block *blockchain.Block) error {
	db := u.Blockchain.Db.Db

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, vin := range tx.Vin {
					updatedOuts := transactions.TXOutputs{}
					outsBytes := b.Get(vin.Txid)
					outs, err := transactions.DeserializeOutputs(outsBytes)
					if err != nil {
						return err
					}

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid)
						if err != nil {
							return err
						}
					} else {
						serializedOut, err := updatedOuts.Serialize()
						if err != nil {
							return err
						}

						err = b.Put(vin.Txid, serializedOut)
						if err != nil {
							return err
						}
					}

				}
			}

			newOutputs := transactions.TXOutputs{}

			newOutputs.Outputs = append(newOutputs.Outputs, tx.Vout...)

			serializedOut, err := newOutputs.Serialize()
			if err != nil {
				return err
			}

			err = b.Put(tx.ID, serializedOut)
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})

	return err
}

func NewUTXOTransaction(from, to string, amount int, UTXOSet *UTXOSet) (*transactions.Transaction, error) {
	var inputs []transactions.TXInput
	var outputs []transactions.TXOutput

	if from == to {
		return nil, fmt.Errorf("you cannot send coins to yourself")
	}

	wallets, err := wallet.NewWallets()
	if err != nil {
		return nil, err
	}

	ws, err := wallets.GetWallet(from)
	if err != nil {
		return nil, err
	}

	pubKeyHash, err := wallet.HashPubKey(ws.PublicKey)
	if err != nil {
		return nil, err
	}

	acc, validOutputs, err := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)
	if err != nil {
		return nil, err
	}

	if acc < amount {
		return nil, fmt.Errorf("not enough funds")
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			return nil, err
		}

		for _, out := range outs {
			input := transactions.TXInput{Txid: txID, Vout: out, Signature: nil, PubKey: ws.PublicKey}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs, *transactions.NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *transactions.NewTXOutput(acc-amount, from)) // a change
	}

	tx := transactions.Transaction{ID: nil, Vin: inputs, Vout: outputs}
	txHash, err := tx.Hash()
	if err != nil {
		return nil, err
	}
	tx.ID = txHash
	UTXOSet.Blockchain.SignTransaction(&tx, ws.PrivateKey)

	return &tx, nil
}
