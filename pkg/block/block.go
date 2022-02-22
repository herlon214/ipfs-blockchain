package block

import (
	"bytes"
	"crypto/sha256"
)

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
}

func New(data string, prevHash []byte) *Block {
	b := Block{
		Data:     []byte(data),
		PrevHash: prevHash,
	}
	b.DeriveHash()

	return &b
}

func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(info)

	b.Hash = hash[:]
}
