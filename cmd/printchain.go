package cmd

import (
	"amdzy/gochain/pkg/blockchain"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var printChainCmd = &cobra.Command{
	Use:   "printchain",
	Short: "print all the blocks of the blockchain",
	Long:  "print all the blocks of the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		bc, err := blockchain.NewBlockchain()
		if err != nil {
			log.Fatal("failed to init blockchain")
		}
		defer bc.CloseDB()

		bci := bc.Iterator()

		for {
			block, err := bci.Next()
			if err != nil {
				log.Fatal("failed to get block")
			}

			fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
			fmt.Printf("Data: %s\n", block.Data)
			fmt.Printf("Hash: %x\n", block.Hash)
			pow := blockchain.NewProofOfWork(block)
			fmt.Printf("Valid: %t\n", pow.Validate())
			fmt.Println()

			if len(block.PrevBlockHash) == 0 {
				break
			}
		}
	},
}
