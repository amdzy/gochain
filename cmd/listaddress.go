package cmd

import (
	"amdzy/gochain/pkg/wallet"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var listAddressesCmd = &cobra.Command{
	Use:   "listaddresses",
	Short: "List addresses",
	Long:  "List addresses",
	Run: func(cmd *cobra.Command, args []string) {
		wallets, err := wallet.NewWallets()
		if err != nil {
			log.Fatal(err)
		}

		addresses := wallets.GetAddresses()

		for _, address := range addresses {
			fmt.Println(address)
		}
	},
}
