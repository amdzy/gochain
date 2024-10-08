package cmd

import (
	"amdzy/gochain/pkg/blockchain"
	"amdzy/gochain/pkg/transactions"
	"amdzy/gochain/pkg/utxo"
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
		UTXOSet := utxo.UTXOSet{Blockchain: bc}
		defer bc.CloseDB()

		tx, err := utxo.NewUTXOTransaction(sendFrom, sendTo, sendAmount, &UTXOSet)
		if err != nil {
			log.Fatal(err)
		}

		coinbaseTx, err := transactions.NewCoinbaseTX(sendFrom, "")
		if err != nil {
			log.Fatal(err)
		}

		block, err := bc.MineBlock([]*transactions.Transaction{tx, coinbaseTx})
		if err != nil {
			log.Fatal(err)
		}

		err = UTXOSet.Update(block)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Success!")
	},
}
