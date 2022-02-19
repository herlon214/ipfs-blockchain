package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/herlon214/ipfs-blockchain/pkg/channel"
	"github.com/herlon214/ipfs-blockchain/pkg/data"
	"github.com/libp2p/go-libp2p-core/crypto"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := flag.Int("p", 4005, "Listen port number")
	keyFile := flag.String("k", "", "Private key file")
	destAddr := flag.String("d", "", "Destination address")

	flag.Parse()

	//Read the key file
	keyBytes, err := os.ReadFile(*keyFile)
	if err != nil {
		panic(err)
	}

	// Unmarshal private key
	prvKey, err := crypto.UnmarshalPrivateKey(keyBytes)
	if err != nil {
		panic(err)
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *port))
	if err != nil {
		panic(err)
	}

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	currentHost, err := libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Started node...")
	log.Printf("Current node /ip4/127.0.0.1/tcp/%v/p2p/%s'\n", *port, currentHost.ID().Pretty())
	for _, addr := range currentHost.Addrs() {
		fmt.Println(addr.String())
	}

	fmt.Println("----------------------------")
	fmt.Println("Starting ipfs node...")
	dataLayer, err := data.New(ctx)
	if err != nil {
		panic(err)
	}

	// Upload a new block
	fmt.Println("Adding random file")

	tmpFile, err := createRandomFile()
	if err != nil {
		panic(err)
	}

	err = dataLayer.UploadBlock(ctx, tmpFile)
	if err != nil {
		panic(err)
	}

	// Connect to destination
	if *destAddr != "" {
		fmt.Println("Connecting to", *destAddr)
		err := connectToPeer(ctx, currentHost, *destAddr)
		if err != nil {
			panic(err)
		}
	}

	// Add all the current files to ipfs
	err = filepath.Walk("./blocks", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		return dataLayer.UploadBlock(ctx, path)
	})

	// create a new PubSub service using the GossipSub router
	ps, err := pubsub.NewGossipSub(ctx, currentHost)
	if err != nil {
		panic(err)
	}

	_, err = channel.NewBlockChannel(ctx, ps, currentHost.ID(), dataLayer.DownloadBlock)
	if err != nil {
		panic(err)
	}

	fmt.Println("Waiting...")
	// Wait forever
	select {}

}

func createRandomFile() (string, error) {
	file, err := os.CreateTemp("./blocks", "block_")
	if err != nil {
		panic(err)
	}

	file.WriteString(fmt.Sprintf("Hello World!!! %d", rand.Int()))

	return file.Name(), nil
}

func connectToPeer(ctx context.Context, h host.Host, destination string) error {
	// Turn the destination into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(destination)
	if err != nil {
		log.Println(err)
		return err
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Println(err)
		return err
	}

	// Add the destination's peer multiaddress in the peerstore.
	// This will be used during connection and stream creation by libp2p.
	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	err = h.Connect(ctx, *info)
	if err != nil {
		return err
	}

	return nil
}
