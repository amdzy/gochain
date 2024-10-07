package wallet

import (
	"crypto/x509"
	"encoding/pem"
)

type walletEncoded struct {
	PrivateKey string
	PublicKey  []byte
}

func encodeWallet(wallet *Wallet) walletEncoded {
	x509Encoded, _ := x509.MarshalECPrivateKey(&wallet.PrivateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	encWallet := walletEncoded{PrivateKey: string(pemEncoded), PublicKey: wallet.PublicKey}

	return encWallet
}

func decodeWallet(encWallet walletEncoded) Wallet {
	block, _ := pem.Decode([]byte(encWallet.PrivateKey))
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	return Wallet{PrivateKey: *privateKey, PublicKey: encWallet.PublicKey}
}
