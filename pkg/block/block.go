package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"

	"github.com/herlon214/ipfs-blockchain/pkg/transaction"
)

type Block struct {
	Hash     []byte
	PrevHash []byte
	Nonce    int

	Transactions []*transaction.Transaction
}

func New(txs []*transaction.Transaction, prevHash []byte) *Block {
	b := Block{
		PrevHash:     prevHash,
		Nonce:        0,
		Transactions: txs,
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

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
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
