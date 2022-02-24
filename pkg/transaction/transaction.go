package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

type Transaction struct {
	ID      []byte
	Inputs  []Input
	Outputs []Output
}

func New(inputs []Input, outputs []Output) *Transaction {
	tx := Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}

	err := tx.SetId()
	if err != nil {
		panic(err)
	}

	return &tx
}

func (tx *Transaction) SetId() error {
	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	if err != nil {
		return err
	}

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]

	return nil
}

func (tx *Transaction) String() string {
	result := fmt.Sprintf("%x\n", tx.ID)

	if tx.IsCoinBase() {
		result += "Coinbase transaction\n"
		result += tx.Outputs[0].String()
	} else {
		for _, input := range tx.Inputs {
			result += input.String() + "\n"
		}
		for _, output := range tx.Outputs {
			result += output.String() + "\n"
		}
	}

	return result
}

func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func CoinBase(to string, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txin := Input{
		ID:  []byte{},
		Out: -1,
		Sig: data,
	}
	txout := Output{
		Value:  100,
		PubKey: to,
	}

	return New([]Input{txin}, []Output{txout})
}
