package main

import (
	"fmt"

	"github.com/herlon214/ipfs-blockchain/pkg/chain"
)

func main() {
	blockChain := chain.New()

	blockChain.AddBlock("First")
	blockChain.AddBlock("Second")
	blockChain.AddBlock("Third")

	for _, block := range blockChain.Blocks {
		fmt.Printf("Previous hash: %x\n", block.PrevHash)
		fmt.Printf("Data in block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
	}
}
