package chain

import (
	"bytes"
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/herlon214/ipfs-blockchain/pkg/block"
)

const (
	dbPath = "./blocks"
)

type Chain struct {
	LastHash []byte

	Database *badger.DB
}

func New() *Chain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		itemResult, err := txn.Get([]byte("lh"))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				genesis := block.New("Genesis", []byte{})
				genesis.DeriveHash()

				err = txn.Set(genesis.Hash, genesis.Serialize())
				if err != nil {
					return err
				}

				err := txn.Set([]byte("lh"), genesis.Hash)
				if err != nil {
					return err
				}

				lastHash = genesis.Hash

				return nil
			}

			return err
		}

		return itemResult.Value(func(val []byte) error {
			lastHash = val

			return nil
		})
	})

	return &Chain{
		Database: db,
		LastHash: lastHash,
	}
}

func (c *Chain) AddBlock(data string) {
	var lastHash []byte

	err := c.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			lastHash = val

			return nil
		})

	})
	if err != nil {
		panic(err)
	}

	newBlock := block.New(data, lastHash)
	newBlock.DeriveHash()

	err = c.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}

		c.LastHash = newBlock.Hash

		return txn.Set([]byte("lh"), newBlock.Hash)

	})
	if err != nil {
		panic(err)
	}
}

func (c *Chain) PrintBlocks() {
	nextHash := c.LastHash

	for {
		if bytes.Compare(nextHash, []byte{}) == 0 {
			break
		}

		err := c.Database.View(func(txn *badger.Txn) error {
			item, err := txn.Get(nextHash)
			if err != nil {
				return err
			}

			err = item.Value(func(val []byte) error {
				currentBlock := block.Deserialize(val)

				fmt.Println("------------------------------------------")
				fmt.Printf("Previous hash: %x\n", currentBlock.PrevHash)
				fmt.Printf("Data in block: %s\n", currentBlock.Data)
				fmt.Printf("Hash: %x\n", currentBlock.Hash)

				nextHash = currentBlock.PrevHash

				return nil
			})

			return err
		})
		if err != nil {
			panic(err)
		}

	}

}
