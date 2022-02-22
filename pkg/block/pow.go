package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
)

const Difficulty = 12

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty))

	return &ProofOfWork{b, target}
}

func (p *ProofOfWork) InitData(nonce int) []byte {
	return bytes.Join(
		[][]byte{
			p.Block.PrevHash,
			p.Block.Data,
		},
		[]byte{},
	)
}

func (p *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0

	for nonce < math.MaxInt64 {
		data := p.InitData(nonce)
		hash := sha256.Sum256(data)

		fmt.Printf("\r%x", hash)
	}
}

func ToHex(num int64) ([]byte, error) {
	buff := bytes.NewBuffer([]byte{})
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
