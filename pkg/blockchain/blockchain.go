package blockchain

import (
	"amdzy/gochain/pkg/transactions"
	"encoding/hex"
	"fmt"
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

func (bc *Blockchain) FindUnspentTransactions(address string) []transactions.Transaction {
	var unspentTXs []transactions.Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			return nil
		}

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

func (bc *Blockchain) FindUTXO(address string) []transactions.TXOutput {
	var UTXOs []transactions.TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
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

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) (*transactions.Transaction, error) {
	var inputs []transactions.TXInput
	var outputs []transactions.TXOutput

	if from == to {
		return nil, fmt.Errorf("you cannot send coins to yourself")
	}

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

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
			input := transactions.TXInput{Txid: txID, Vout: out, ScriptSig: from}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs, transactions.TXOutput{Value: amount, ScriptPubKey: to})
	if acc > amount {
		outputs = append(outputs, transactions.TXOutput{Value: acc - amount, ScriptPubKey: from}) // a change
	}

	tx := transactions.Transaction{ID: nil, Vin: inputs, Vout: outputs}
	tx.SetID()

	return &tx, nil
}
