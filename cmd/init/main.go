package main

import (
	"crypto/rand"
	"flag"
	"os"

	"github.com/libp2p/go-libp2p-core/crypto"
)

func main() {
	dest := flag.String("f", "", "Destination private key file")

	flag.Parse()

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		panic(err)
	}

	// Marshal the key
	key, err := crypto.MarshalPrivateKey(prvKey)
	if err != nil {
		panic(err)
	}

	// Save the private key into a file
	err = os.WriteFile(*dest, key, 0600)
	if err != nil {
		panic(err)
	}

}
