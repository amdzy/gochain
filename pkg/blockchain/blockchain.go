package blockchain

import (
	"amdzy/gochain/pkg/transactions"
	"amdzy/gochain/pkg/wallet"
	"bytes"
	"crypto/ecdsa"
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

	for _, tx := range transactions {
		verified, err := bc.VerifyTransaction(tx)
		if err != nil {
			return err
		}

		if !verified {
			return fmt.Errorf("invalid transaction")
		}
	}

	block, err := NewBlock(transactions, tip)
	if err != nil {
		return err
	}

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

func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []transactions.Transaction {
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

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
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

func (bc *Blockchain) FindUTXO(pubKeyHash []byte) []transactions.TXOutput {
	var UTXOs []transactions.TXOutput
	unspentTransactions := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
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

	acc, validOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)

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
	bc.SignTransaction(&tx, ws.PrivateKey)

	return &tx, nil
}
