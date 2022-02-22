package data

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	config "github.com/ipfs/go-ipfs-config"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/core/node/libp2p"
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	icore "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"golang.org/x/net/context"
)

type Data struct {
	mx     sync.Mutex
	blocks map[string]string
	ipfs   icore.CoreAPI
}

func New(ctx context.Context) (*Data, error) {
	if err := createBlocksFolder(); err != nil {
		return nil, err
	}

	if err := setupPlugins(""); err != nil {
		return nil, err
	}

	// Create a Temporary Repo
	repoPath, err := createTempRepo()
	if err != nil {
		return nil, fmt.Errorf("failed to create temp repo: %s", err)
	}

	// Spawning an ephemeral IPFS node
	ipfsNode, err := createNode(ctx, repoPath)
	if err != nil {
		return nil, err
	}

	return &Data{
		blocks: make(map[string]string, 0),
		ipfs:   ipfsNode,
	}, nil
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
		Routing: libp2p.DHTOption, // This option sets the node to be a full DHT node (both fetching and storing DHT Records)
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

func (d *Data) UploadBlock(ctx context.Context, tmpFile string) error {
	someFile, err := getUnixfsNode(tmpFile)
	if err != nil {
		return err
	}

	cidFile, err := d.ipfs.Unixfs().Add(ctx, someFile)
	if err != nil {
		return err
	}

	d.mx.Lock()
	d.blocks[cidFile.String()] = tmpFile
	d.mx.Unlock()

	fmt.Printf("Added file to IPFS with CID %s to path %s\n", cidFile.String(), tmpFile)

	return nil
}

func (d *Data) DownloadBlock(ctx context.Context, fileCid string, filePath string) error {
	if _, ok := d.blocks[fileCid]; ok {
		return nil
	}

	ipfsPath := path.New(fileCid)

	rootNodeFile, err := d.ipfs.Unixfs().Get(ctx, ipfsPath)
	if err != nil {
		return err
	}

	err = files.WriteTo(rootNodeFile, filePath)
	if err != nil {
		return err
	}

	d.mx.Lock()
	fmt.Println("Downloaded block", fileCid, filePath)
	d.blocks[fileCid] = filePath
	d.mx.Unlock()

	return nil
}
