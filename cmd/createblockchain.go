package cmd

import (
	"amdzy/gochain/pkg/blockchain"
	"amdzy/gochain/pkg/utxo"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func NewCreateBlockchainCommand() *cobra.Command {
	var address string

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

			UTXOSet := utxo.UTXOSet{Blockchain: bc}
			UTXOSet.ReIndex()

			fmt.Println("Done!")
		},
	}

	createBlockChainCmd.Flags().StringVarP(&address, "address", "a", "", "The address to send genesis block reward to")
	cobra.MarkFlagRequired(createBlockChainCmd.Flags(), "address")

	return createBlockChainCmd
}
