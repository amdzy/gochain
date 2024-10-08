package transactions

import (
	"amdzy/gochain/utils"
	"bytes"

	"github.com/vmihailenco/msgpack/v5"
)

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := utils.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}

func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

type TXOutputs struct {
	Outputs []TXOutput
}

func (outs TXOutputs) Serialize() ([]byte, error) {
	return msgpack.Marshal(outs)
}

func DeserializeOutputs(data []byte) (TXOutputs, error) {
	var outputs TXOutputs

	err := msgpack.Unmarshal(data, &outputs)
	if err != nil {
		return TXOutputs{}, err
	}

	return outputs, nil
}
