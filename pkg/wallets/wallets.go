package wallets

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"os"
)

type Wallets struct {
	FilePath string
	Items    map[string]*Wallet
}

func init() {
	gob.Register(map[string]*Wallet{})
	gob.Register(elliptic.P256())
}

func New(filePath string) (*Wallets, error) {
	ws := &Wallets{
		FilePath: filePath,
		Items:    make(map[string]*Wallet, 0),
	}

	if err := ws.Save(); err != nil {
		return nil, err
	}

	return ws, nil
}

func (ws *Wallets) Save() error {
	var content bytes.Buffer

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		return err
	}

	return os.WriteFile(ws.FilePath, content.Bytes(), 0644)
}

func (ws *Wallets) Wallet(address string) *Wallet {
	if wallet, ok := ws.Items[address]; ok {
		return wallet
	}

	return nil
}

func (ws *Wallets) NewWallet() (*Wallet, error) {
	wallet, err := NewWallet()
	if err != nil {
		return nil, err
	}

	ws.Items[string(wallet.Address())] = wallet

	return wallet, nil
}

func Load(filePath string) (*Wallets, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}

	var wallets Wallets
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(&wallets)
	if err != nil {
		return nil, err
	}

	return &wallets, nil
}
