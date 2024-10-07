package wallet

import (
	"amdzy/gochain/utils"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

const addressChecksumLen = 4
const version = byte(0x00)
const walletFile = "wallets.dat"

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func (w Wallet) GetAddress() ([]byte, error) {
	pubKeyHash, err := HashPubKey(w.PublicKey)
	if err != nil {
		return nil, err
	}

	versionPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionPayload)

	fullPayload := append(versionPayload, checksum...)
	address := utils.Base58Encode(fullPayload)

	return address, nil
}

func HashPubKey(pubKey []byte) ([]byte, error) {
	publicSha265 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSha265[:])
	if err != nil {
		return nil, err
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160, nil
}

func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

func ValidateAddress(address string) bool {
	pubKeyHash := utils.Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Equal(actualChecksum, targetChecksum)
}

func NewWallet() (*Wallet, error) {
	private, public, err := NewKeyPair()
	if err != nil {
		return nil, err
	}

	wallet := &Wallet{private, public}
	return wallet, nil
}

func NewKeyPair() (ecdsa.PrivateKey, []byte, error) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return ecdsa.PrivateKey{}, nil, err
	}
	public := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, public, nil
}
