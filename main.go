package main

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/herlon214/ipfs-blockchain/pkg/chain"
	"github.com/libp2p/go-libp2p-core/crypto"
)

func main() {
	blockChain := chain.New()
	defer blockChain.Database.Close()

	// blockChain.AddBlock("First")
	// blockChain.AddBlock("Second")
	// blockChain.AddBlock("Third")

	blockChain.PrintBlocks()

	herlon()
}

func herlon() {
	privKey, err := readKey("herlon.key")
	if err != nil {
		panic(err)
	}

	pubKey := privKey.GetPublic()
	if err != nil {
		panic(err)
	}

	fmt.Println("Key hash", KeyHash(pubKey))

}

func KeyHash(key crypto.Key) string {
	pubKeyRaw, err := key.Raw()
	if err != nil {
		panic(err)
	}

	pubHash := sha256.Sum256(pubKeyRaw)

	return fmt.Sprintf("%x", pubHash[:])
}

func readKey(name string) (crypto.PrivKey, error) {
	//Read the key file
	keyBytes, err := os.ReadFile(name)
	if err != nil {
		panic(err)
	}

	// Unmarshal private key
	return crypto.UnmarshalPrivateKey(keyBytes)
}
