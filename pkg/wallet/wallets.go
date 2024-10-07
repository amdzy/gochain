package wallet

import (
	"fmt"
	"os"

	"github.com/vmihailenco/msgpack/v5"
)

type Wallets struct {
	Wallets map[string]*Wallet
}

func (ws *Wallets) CreateWallet() (string, error) {
	wallet, err := NewWallet()
	if err != nil {
		return "", err
	}

	addr, err := wallet.GetAddress()
	if err != nil {
		return "", err
	}
	address := string(addr)

	ws.Wallets[address] = wallet

	return address, nil
}

func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

func (ws *Wallets) GetWallet(address string) (Wallet, error) {
	wallet, ok := ws.Wallets[address]

	if !ok {
		return Wallet{}, fmt.Errorf("account not found")
	}

	return *wallet, nil
}

func (ws *Wallets) LoadFromFile() error {
	_, err := os.Stat(walletFile)
	if os.IsNotExist(err) {
		return nil
	}

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		return err
	}

	var encWallets []walletEncoded
	err = msgpack.Unmarshal(fileContent, &encWallets)
	if err != nil {
		return err
	}

	var wallets = Wallets{
		Wallets: map[string]*Wallet{},
	}

	for _, encWs := range encWallets {
		wallet := decodeWallet(encWs)

		addr, err := wallet.GetAddress()
		if err != nil {
			return err
		}
		address := string(addr)

		wallets.Wallets[address] = &wallet
	}

	ws.Wallets = wallets.Wallets

	return nil
}

func (ws *Wallets) SaveToFile() error {
	var encWallets []walletEncoded

	for _, wallet := range ws.Wallets {
		encWallets = append(encWallets, encodeWallet(wallet))
	}

	b, err := msgpack.Marshal(encWallets)
	if err != nil {
		return err
	}

	os.WriteFile(walletFile, b, 0644)

	return nil
}

func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile()
	if err != nil {
		return nil, err
	}

	return &wallets, nil
}
