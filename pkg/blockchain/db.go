package blockchain

import (
	"amdzy/gochain/pkg/transactions"
	"fmt"
	"log"

	bolt "go.etcd.io/bbolt"
)

const blocksBucket = "blocks"

type DB struct {
	db *bolt.DB
}

func (db *DB) GetTip() ([]byte, error) {
	var tip []byte

	err := db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		tip = b.Get([]byte("l"))

		return nil
	})

	return tip, err
}

func (db *DB) AddBlock(block *Block) error {
	err := db.db.Update(func(tx *bolt.Tx) error {
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

	err := db.db.View(func(tx *bolt.Tx) error {
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
	db.db.Close()
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

		genesis := NewGenesisBlock(coinbaseTx)
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

	return &DB{db: db}, err
}

func ConnectDB() (*DB, error) {
	db, err := bolt.Open("blockchain.db", 0600, nil)
	if err != nil {
		log.Fatal("failed to connect to db")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			return fmt.Errorf("no existing blockchain found. Create one first")
		}

		return nil
	})

	return &DB{db: db}, err
}
