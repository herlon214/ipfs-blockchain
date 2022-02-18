package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	config "github.com/ipfs/go-ipfs-config"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	ipfsp2p "github.com/ipfs/go-ipfs/core/node/libp2p"
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	icore "github.com/ipfs/interface-go-ipfs-core"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := flag.Int("p", 4005, "Listen port number")
	keyFile := flag.String("k", "", "Private key file")

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
	host, err := libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Started node...")
	fmt.Println(host.ID().String())
	for _, addr := range host.Addrs() {
		fmt.Println(addr.String())
	}

	fmt.Println("----------------------------")
	fmt.Println("Starting ipfs node...")
	ipfs, err := startIPFS(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println("Adding random file")

	tmpFile, err := createRandomFile()
	if err != nil {
		panic(err)
	}

	someFile, err := getUnixfsNode(tmpFile)
	if err != nil {
		panic(fmt.Errorf("Could not get File: %s", err))
	}

	cidFile, err := ipfs.Unixfs().Add(ctx, someFile)
	if err != nil {
		panic(fmt.Errorf("Could not add File: %s", err))
	}

	fmt.Printf("Added file to IPFS with CID %s\n", cidFile.String())

	fmt.Println("Waiting...")
	// Wait forever
	select {}

}

func createBlocksFolder() error {
	// Return if already exists
	if _, err := os.Stat("blocks"); !os.IsNotExist(err) {
		return nil
	}

	err := os.Mkdir("blocks", 0600)
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

	err := os.Mkdir(folderName, 0600)
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

	file.WriteString("Hello World!!!")

	return file.Name(), nil
}
