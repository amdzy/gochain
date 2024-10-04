package cmd

import (
	"amdzy/gochain/pkg/blockchain"
	"log"

	"github.com/spf13/cobra"
)

var newBlockData string
var addBlockCmd = &cobra.Command{
	Use:   "addblock",
	Short: "--data BLOCK_DATA - add a block to the blockchain",
	Long:  "--data BLOCK_DATA - add a block to the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		bc, err := blockchain.NewBlockchain()
		if err != nil {
			log.Fatal("failed to init blockchain")
		}
		defer bc.CloseDB()

		// err = bc.AddBlock(newBlockData)
		// if err != nil {
		// 	log.Fatal("failed to add block")
		// }
	},
}
