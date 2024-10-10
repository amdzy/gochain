package cmd

import (
	"amdzy/gochain/pkg/blockchain"
	"amdzy/gochain/pkg/utxo"
	"amdzy/gochain/utils"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func NewGetBalanceCommand() *cobra.Command {
	var address string

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
			UTXOSet := utxo.UTXOSet{Blockchain: bc}

			balance := 0
			pubKeyHash := utils.Base58Decode([]byte(address))
			pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
			UTXOs, err := UTXOSet.FindUTXO(pubKeyHash)
			if err != nil {
				log.Fatal(err)
			}

			for _, out := range UTXOs {
				balance += out.Value
			}

			fmt.Printf("Balance of '%s': %d\n", address, balance)
		},
	}

	getBalanceCmd.Flags().StringVarP(&address, "address", "a", "", "The address to send genesis block reward to")
	cobra.MarkFlagRequired(getBalanceCmd.Flags(), "address")

	return getBalanceCmd
}
