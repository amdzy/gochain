package cmd

import (
	"amdzy/gochain/pkg/blockchain"
	"amdzy/gochain/pkg/utxo"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var reIndexUTXoCmd = &cobra.Command{
	Use:   "reindexutxo",
	Short: "re-index UTXOs",
	Long:  "re-index UTXOs",
	Run: func(cmd *cobra.Command, args []string) {
		bc, err := blockchain.NewBlockchain()
		if err != nil {
			log.Fatal(err)
		}
		defer bc.CloseDB()
		UTXOSet := utxo.UTXOSet{Blockchain: bc}

		err = UTXOSet.ReIndex()
		if err != nil {
			log.Fatal(err)
		}

		count, err := UTXOSet.CountTransactions()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
	},
}
