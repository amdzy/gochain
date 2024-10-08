package transactions

import (
	"amdzy/gochain/pkg/wallet"
	"bytes"
)

type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash, err := wallet.HashPubKey(in.PubKey)
	if err != nil {
		return false
	}

	return bytes.Equal(lockingHash, pubKeyHash)
}
