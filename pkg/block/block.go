package block

import (
	"bytes"
	"encoding/gob"
)

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

func New(data string, prevHash []byte) *Block {
	b := Block{
		Data:     []byte(data),
		PrevHash: prevHash,
		Nonce:    0,
	}

	b.DeriveHash()

	return &b
}

func (b *Block) DeriveHash() {
	pow := NewProof(b)
	nonce, hash := pow.Run()

	b.Hash = hash
	b.Nonce = nonce
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer

	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)
	if err != nil {
		panic(err)
	}

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)
	if err != nil {
		panic(err)
	}

	return &block
}
