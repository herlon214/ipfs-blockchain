package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/herlon214/ipfs-blockchain/pkg/channel"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"

	config "github.com/ipfs/go-ipfs-config"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	ipfsp2p "github.com/ipfs/go-ipfs/core/node/libp2p"
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	icore "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/libp2p/go-libp2p-core/crypto"
)

// DiscoveryInterval is how often we re-publish our mDNS records.
const DiscoveryInterval = time.Hour

// DiscoveryServiceTag is used in our mDNS advertisements to discover other chat peers.
const DiscoveryServiceTag = "blockchain"

type Message struct {
	PeerID string `json:"peerId"`
	Type   string `json:"type"`
	Data   string `json:"data"`
}

type Blocks struct {
	Items map[string]string `json:"items"`
}

var streams = make(map[string]*bufio.ReadWriter, 0)
var connectedPeers = make([]string, 0)
var mx sync.Mutex
var blocks = make(map[string]string, 0) // cid -> filename
var ipfs icore.CoreAPI
var currentHost host.Host
var ps *pubsub.PubSub

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := flag.Int("p", 4005, "Listen port number")
	keyFile := flag.String("k", "", "Private key file")
	destAddr := flag.String("d", "", "Destination address")

	flag.Parse()

	// Create blocks folder
	if err := createBlocksFolder(); err != nil {
		panic(err)
	}

	// Read the key file
	data, err := os.ReadFile(*keyFile)
	if err != nil {
		panic(err)
	}

	// Unmarshal private key
	prvKey, err := crypto.UnmarshalPrivateKey(data)
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
	currentHost, err = libp2p.New(
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

	// currentHost.SetStreamHandler("/blockchain/1.0.0", handleStream)

	fmt.Println("----------------------------")
	fmt.Println("Starting ipfs node...")
	ipfs, err = startIPFS(ctx)
	if err != nil {
		panic(err)
	}

	// Upload a new block
	fmt.Println("Adding random file")

	tmpFile, err := createRandomFile()
	if err != nil {
		panic(err)
	}

	err = uploadBlock(ctx, ipfs, tmpFile)
	if err != nil {
		panic(err)
	}

	// Connect to destination
	if *destAddr != "" {
		fmt.Println("Connecting to", *destAddr)
		err := startPeerAndConnect(ctx, currentHost, *destAddr)
		if err != nil {
			panic(err)
		}

		// mx.Lock()
		// connectedPeers = append(connectedPeers, *destAddr)
		// streams[*destAddr] = rw
		// mx.Unlock()

		// requestPeers(rw)
	}

	// Add all the current files to ipfs
	err = filepath.Walk("./blocks", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		return uploadBlock(ctx, ipfs, path)
	})

	// create a new PubSub service using the GossipSub router
	ps, err = pubsub.NewGossipSub(ctx, currentHost)
	if err != nil {
		panic(err)
	}

	// Blocks topic
	blockChan, err := channel.NewBlockChannel(ctx, ps, currentHost.ID(), downloadBlock)
	if err != nil {
		panic(err)
	}

	//go printPeers()
	go broadcastBlocks(blockChan)

	// if err := setupDiscovery(currentHost); err != nil {
	// 	panic(err)
	// }

	fmt.Println("Waiting...")
	// Wait forever
	select {}

}

func printPeers() {
	for {
		time.Sleep(time.Second * 5)
		fmt.Println(blocks)
	}
}

func broadcastBlocks(blockChan *channel.BlockChannel) {
	for {
		time.Sleep(time.Second * 5)

		peers := currentHost.Network().Peers()
		fmt.Println("Broadcasting blocks to", len(peers), "peers")

		err := blockChan.BroadcastBlocks(blocks)
		if err != nil {
			fmt.Println(err)
		}

	}
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound connects to peers discovered via mDNS. Once they're connected,
// the PubSub system will automatically start interacting with them if they also
// support PubSub.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID.Pretty())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
}

// setupDiscovery creates an mDNS discovery service and attaches it to the libp2p Host.
// This lets us automatically discover peers on the same LAN and connect to them.
func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, DiscoveryServiceTag, &discoveryNotifee{h: h})
	return s.Start()
}

func uploadBlock(ctx context.Context, ipfs icore.CoreAPI, tmpFile string) error {
	someFile, err := getUnixfsNode(tmpFile)
	if err != nil {
		return err
	}

	cidFile, err := ipfs.Unixfs().Add(ctx, someFile)
	if err != nil {
		return err
	}

	mx.Lock()
	blocks[cidFile.String()] = tmpFile
	mx.Unlock()

	fmt.Printf("Added file to IPFS with CID %s to path %s\n", cidFile.String(), tmpFile)

	return nil
}

func downloadBlock(ctx context.Context, fileCid string, filePath string) error {
	if _, ok := blocks[fileCid]; ok {
		return nil
	}

	ipfsPath := path.New(fileCid)

	rootNodeFile, err := ipfs.Unixfs().Get(ctx, ipfsPath)
	if err != nil {
		return err
	}

	err = files.WriteTo(rootNodeFile, filePath)
	if err != nil {
		return err
	}

	mx.Lock()
	fmt.Println("Downloaded block", fileCid, filePath)
	blocks[fileCid] = filePath
	mx.Unlock()

	return nil
}

func startPeerAndConnect(ctx context.Context, h host.Host, destination string) error {
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

func createBlocksFolder() error {
	// Return if already exists
	if _, err := os.Stat("blocks"); !os.IsNotExist(err) {
		return nil
	}

	err := os.Mkdir("blocks", 0755)
	if err != nil {
		return err
	}

	return nil
}

func getUnixfsNode(path string) (files.Node, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	f, err := files.NewSerialFile(path, false, st)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func createNode(ctx context.Context, repoPath string) (icore.CoreAPI, error) {
	// Open the repo
	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		return nil, err
	}

	// Construct the node
	nodeOptions := &core.BuildCfg{
		Online:  true,
		Routing: ipfsp2p.DHTOption, // This option sets the node to be a full DHT node (both fetching and storing DHT Records)
		// Routing: libp2p.DHTClientOption, // This option sets the node to be a client DHT node (only fetching records)
		Repo: repo,
	}

	node, err := core.NewNode(ctx, nodeOptions)
	if err != nil {
		return nil, err
	}

	// Attach the Core API to the constructed node
	return coreapi.NewCoreAPI(node)
}

func startIPFS(ctx context.Context) (icore.CoreAPI, error) {
	if err := setupPlugins(""); err != nil {
		return nil, err
	}

	// Create a Temporary Repo
	repoPath, err := createTempRepo()
	if err != nil {
		return nil, fmt.Errorf("failed to create temp repo: %s", err)
	}

	// Spawning an ephemeral IPFS node
	return createNode(ctx, repoPath)
}

func setupPlugins(externalPluginsPath string) error {
	// Load any external plugins if available on externalPluginsPath
	plugins, err := loader.NewPluginLoader(filepath.Join(externalPluginsPath, "plugins"))
	if err != nil {
		return fmt.Errorf("error loading plugins: %s", err)
	}

	// Load preloaded and external plugins
	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	return nil
}

func createTempRepo() (string, error) {
	folderName := "block-repo"

	// Return if already exists
	if _, err := os.Stat(folderName); !os.IsNotExist(err) {
		return folderName, nil
	}

	err := os.Mkdir(folderName, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to get temp dir: %s", err)
	}

	// Create a config with default options and a 2048 bit key
	cfg, err := config.Init(ioutil.Discard, 2048)
	if err != nil {
		return "", err
	}

	// Create the repo with the config
	err = fsrepo.Init(folderName, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to init ephemeral node: %s", err)
	}

	return folderName, nil
}

func createRandomFile() (string, error) {
	file, err := os.CreateTemp("./blocks", "block_")
	if err != nil {
		panic(err)
	}

	file.WriteString(fmt.Sprintf("Hello World!!! %d", rand.Int()))

	return file.Name(), nil
}

func handleStream(s network.Stream) {
	remoteConn := fmt.Sprintf("%s/p2p/%s", s.Conn().RemoteMultiaddr().String(), s.Conn().RemotePeer().String())
	fmt.Println("Received stream from", remoteConn)

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	mx.Lock()
	connectedPeers = append(connectedPeers, remoteConn)
	streams[remoteConn] = rw
	mx.Unlock()

	go readData(rw)

	fmt.Println("Received stream done")
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			// Unmarshal message
			var msg Message
			err := json.Unmarshal([]byte(str), &msg)
			if err != nil {
				panic(err)
			}

			//fmt.Println("---------------------------")
			//fmt.Println("Message received...")
			//fmt.Printf("Type: %s\n", msg.Type)
			//fmt.Printf("Data: %s\n", msg.Data)
			//fmt.Println("---------------------------")
			switch msg.Type {
			case "RequestPeers":
				sendPeers(rw)
			case "Blocks":
				var peerBlocks Blocks
				err := json.Unmarshal([]byte(msg.Data), &peerBlocks)
				if err != nil {
					panic(err)
				}

				fmt.Println("Received", len(peerBlocks.Items))
				fmt.Println(msg.PeerID)

				for cid, blockPath := range peerBlocks.Items {
					err = downloadBlock(context.Background(), cid, blockPath)
					if err != nil {
						panic(err)
					}
				}
			}

		}

	}
}

func broadcastMessage(msg Message) {
	mx.Lock()
	defer mx.Unlock()

	for _, rw := range streams {
		sendMsg(rw, msg)
	}
}

func requestPeers(rw *bufio.ReadWriter) {
	sendMsg(rw, Message{Type: "RequestPeers"})
}

func sendMsg(rw *bufio.ReadWriter, msg Message) {
	msg.PeerID = currentHost.ID().Pretty()
	out, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	if _, err := rw.WriteString(fmt.Sprintf("%s\n", string(out))); err != nil {
		fmt.Println(err)
		return
	}

	if err := rw.Flush(); err != nil {
		fmt.Println(err)
		return
	}
}

func sendPeers(rw *bufio.ReadWriter) {
	//fmt.Println("Sending peers...")
	mx.Lock()
	peers := make([]string, len(connectedPeers))
	for _, peerId := range connectedPeers {
		peers = append(peers, peerId)
	}
	mx.Unlock()

	//fmt.Println(peers)

	for _, peerId := range peers {
		sendMsg(rw, Message{
			Type: "NewPeer",
			Data: peerId,
		})
	}
	//fmt.Println("Done")
}
