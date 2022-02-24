package chain

import (
	"bytes"
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/herlon214/ipfs-blockchain/pkg/block"
	"github.com/herlon214/ipfs-blockchain/pkg/transaction"
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

				coinbaseTx := transaction.CoinBase("abb9af250a0b4c505838b512ff7dc91a944d1d7dcb1c4b5936a02b56894b2e08", "")
				genesis := block.New([]*transaction.Transaction{coinbaseTx}, []byte{})
				genesis.DeriveHash()

				key := bytes.Join([][]byte{[]byte("block-"), genesis.Hash}, []byte{})
				err = txn.Set(key, genesis.Serialize())
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
	if err != nil {
		panic(err)
	}

	return &Chain{
		Database: db,
		LastHash: lastHash,
	}
}

func (c *Chain) AddBlock(txs []*transaction.Transaction) {
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

	newBlock := block.New(txs, lastHash)
	newBlock.DeriveHash()

	err = c.Database.Update(func(txn *badger.Txn) error {
		key := bytes.Join([][]byte{[]byte("block-"), newBlock.Hash}, []byte{})
		err := txn.Set(key, newBlock.Serialize())
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
		if bytes.Equal(nextHash, []byte{}) {
			break
		}

		err := c.Database.View(func(txn *badger.Txn) error {
			key := bytes.Join([][]byte{[]byte("block-"), nextHash}, []byte{})
			item, err := txn.Get(key)
			if err != nil {
				return err
			}

			err = item.Value(func(val []byte) error {
				currentBlock := block.Deserialize(val)

				fmt.Println("------------------------------------------")
				fmt.Printf("Previous hash: %x\n", currentBlock.PrevHash)
				fmt.Printf("Transactions in block: %d\n", len(currentBlock.Transactions))
				for _, tx := range currentBlock.Transactions {
					fmt.Println(tx.String())
				}
				fmt.Printf("Block hash: %x\n", currentBlock.Hash)
				fmt.Println("------------------------------------------")

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
