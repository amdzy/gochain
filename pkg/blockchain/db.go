package blockchain

import (
	"amdzy/gochain/pkg/transactions"
	"fmt"
	"log"

	bolt "go.etcd.io/bbolt"
)

const blocksBucket = "blocks"

type DB struct {
	Db *bolt.DB
}

func (db *DB) GetLastHashAndHeight() ([]byte, int, error) {
	var lastHash []byte
	var lastHeight int

	err := db.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		blockData := b.Get(lastHash)
		block, err := DeserializeBlock(blockData)
		if err != nil {
			return err
		}

		lastHeight = block.Height

		return nil
	})

	return lastHash, lastHeight, err
}

func (db *DB) AddBlock(block *Block) error {
	err := db.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockBytes, err := block.Serialize()
		if err != nil {
			return err
		}

		err = b.Put(block.Hash, blockBytes)
		if err != nil {
			return err
		}

		err = b.Put([]byte("l"), block.Hash)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (db *DB) GetBlock(hash []byte) (*Block, error) {
	var block *Block

	err := db.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(hash)

		var err error
		block, err = DeserializeBlock(encodedBlock)
		if err != nil {
			return err
		}

		return nil
	})

	return block, err
}

func (db *DB) Close() {
	db.Db.Close()
}

func InitDB(address string) (*DB, error) {
	db, err := bolt.Open("blockchain.db", 0600, nil)
	if err != nil {
		log.Fatal("failed to connect to db")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b != nil {
			return fmt.Errorf("blockchain already exists")
		}

		coinbaseTx, err := transactions.NewCoinbaseTX(address, "")
		if err != nil {
			return err
		}

		genesis, err := NewGenesisBlock(coinbaseTx)
		if err != nil {
			return err
		}

		b, err = tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return err
		}

		genesisBytes, err := genesis.Serialize()
		if err != nil {
			return err
		}

		err = b.Put(genesis.Hash, genesisBytes)
		if err != nil {
			return err
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			return err
		}

		return nil
	})

	return &DB{Db: db}, err
}

func ConnectDB() (*DB, error) {
	db, err := bolt.Open("blockchain.db", 0600, nil)
	if err != nil {
		log.Fatal("failed to connect to db")
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			return fmt.Errorf("no existing blockchain found. Create one first")
		}

		return nil
	})

	return &DB{Db: db}, err
}
