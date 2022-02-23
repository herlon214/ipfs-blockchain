package main

import (
	"github.com/herlon214/ipfs-blockchain/pkg/chain"
)

func main() {
	blockChain := chain.New()

	// blockChain.AddBlock("First")
	// blockChain.AddBlock("Second")
	// blockChain.AddBlock("Third")

	blockChain.PrintBlocks()
}
