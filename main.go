package main

import (
	"fmt"
	"os"

	"github.com/herlon214/ipfs-blockchain/cmd"
	"github.com/herlon214/ipfs-blockchain/pkg/wallets"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// blockChain := chain.New()
	// defer blockChain.Database.Close()

	// blockChain.AddBlock("First")
	// blockChain.AddBlock("Second")
	// blockChain.AddBlock("Third")

	// blockChain.PrintBlocks()

	// herlonWallet()
}

func herlonWallet() {
	// Load wallet
	walletFile := "herlon_wallet.dat"
	ws, err := wallets.Load(walletFile)
	if err != nil {
		ws, err = wallets.New(walletFile)
		if err != nil {
			panic(err)
		}
	}

	defer ws.Save()

	if len(ws.Items) == 0 {
		_, err := ws.NewWallet()
		if err != nil {
			panic(err)
		}
	}

	for _, wallet := range ws.Items {
		fmt.Println("----------------------------------")
		fmt.Printf("Public key: %x\n", wallet.PublicKey)
		fmt.Printf("Address: %s\n", wallet.Address())
	}

}
