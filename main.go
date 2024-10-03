package main

import (
	"fmt"
	"log"
)

func main() {
	db, err := InitDB()
	if err != nil {
		log.Fatal("failed to connect to db")
	}
	defer db.Close()

	bc, err := NewBlockchain(db)
	if err != nil {
		log.Fatal("failed to init blockchain")
	}

	err = bc.AddBlock("Send 1 BTC to Ivan")
	if err != nil {
		log.Fatal("failed to add block")
	}

	err = bc.AddBlock("Send 2 more BTC to Ivan")
	if err != nil {
		log.Fatal("failed to add block")
	}

	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			log.Fatal("failed to get block")
		}

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("Valid: %t\n", pow.Validate())
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}
