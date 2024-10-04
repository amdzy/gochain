package transactions

import (
	"crypto/sha256"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

var subsidy = 10

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

func (tx *Transaction) SetID() error {
	enc, err := msgpack.Marshal(tx)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(enc)
	tx.ID = hash[:]

	return nil
}

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func NewCoinbaseTX(to, data string) (*Transaction, error) {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	err := tx.SetID()
	if err != nil {
		return nil, err
	}
	return &tx, nil
}
