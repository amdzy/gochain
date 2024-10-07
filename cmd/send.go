package cmd

import (
	"amdzy/gochain/pkg/blockchain"
	"amdzy/gochain/pkg/transactions"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var sendFrom string
var sendTo string
var sendAmount int

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "--from FROM --to TO --amount AMOUNT - send coins to another address",
	Long:  "--from FROM --to TO --amount AMOUNT - send coins to another address",
	Run: func(cmd *cobra.Command, args []string) {
		if sendAmount <= 0 {
			fmt.Println("Amount can't be less than 0")
			cmd.Help()
			os.Exit(1)
		}

		bc, err := blockchain.NewBlockchain()
		if err != nil {
			log.Fatal(err)
		}
		defer bc.CloseDB()

		tx, err := blockchain.NewUTXOTransaction(sendFrom, sendTo, sendAmount, bc)
		if err != nil {
			log.Fatal(err)
		}

		err = bc.MineBlock([]*transactions.Transaction{tx})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Success!")
	},
}