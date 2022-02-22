package chain

import "github.com/herlon214/ipfs-blockchain/pkg/block"

type Chain struct {
	Blocks []*block.Block
}

func New() *Chain {
	return &Chain{
		Blocks: []*block.Block{block.New("Genesis", []byte(""))}, // Genesis block
	}
}

func (c *Chain) AddBlock(data string) {
	prevBlock := c.Blocks[len(c.Blocks)-1]
	newBlock := block.New(data, prevBlock.Hash)

	c.Blocks = append(c.Blocks, newBlock)
}
