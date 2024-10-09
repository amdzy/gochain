package cmd

import (
	"amdzy/gochain/pkg/blockchain"
	"amdzy/gochain/pkg/server"
	"amdzy/gochain/pkg/utxo"
	"amdzy/gochain/pkg/wallet"
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

		wallets, err := wallet.NewWallets()
		if err != nil {
			log.Panic(err)
		}

		wallet, err := wallets.GetWallet(sendFrom)
		if err != nil {
			log.Panic(err)
		}

		tx, err := utxo.NewUTXOTransaction(&wallet, sendTo, sendAmount, &UTXOSet)
		if err != nil {
			log.Fatal(err)
		}

		server.SendTx(server.KnownNodes[0], tx)

		fmt.Println("Success!")
	},
}
