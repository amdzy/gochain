package cmd

import (
	"amdzy/gochain/pkg/blockchain"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var createBlockChainCmd = &cobra.Command{
	Use:   "createblockchain",
	Short: "-address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS",
	Long:  "-address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS",
	Run: func(cmd *cobra.Command, args []string) {
		bc, err := blockchain.CreateBlockChain(address)
		if err != nil {
			log.Fatal(err)
		}
		defer bc.CloseDB()

		fmt.Println("Done!")
	},
}
