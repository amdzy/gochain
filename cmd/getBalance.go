package cmd

import (
	"amdzy/gochain/pkg/blockchain"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var getBalanceCmd = &cobra.Command{
	Use:   "getbalance",
	Short: "--data BLOCK_DATA - add a block to the blockchain",
	Long:  "--data BLOCK_DATA - add a block to the blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		bc, err := blockchain.NewBlockchain()
		if err != nil {
			log.Fatal(err)
		}
		defer bc.CloseDB()

		balance := 0
		UTXOs := bc.FindUTXO(address)

		for _, out := range UTXOs {
			balance += out.Value
		}

		fmt.Printf("Balance of '%s': %d\n", address, balance)
	},
}
