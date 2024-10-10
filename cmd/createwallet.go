package cmd

import (
	"amdzy/gochain/pkg/wallet"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func NewCreateWalletCommand() *cobra.Command {
	var createWalletCmd = &cobra.Command{
		Use:   "createwallet",
		Short: "create wallet",
		Long:  "create wallet",
		Run: func(cmd *cobra.Command, args []string) {
			wallets, err := wallet.NewWallets()
			if err != nil {
				log.Fatal(err)
			}

			address, err := wallets.CreateWallet()
			if err != nil {
				log.Fatal(err)
			}

			err = wallets.SaveToFile()
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Your new address: %s\n", address)
		},
	}

	return createWalletCmd
}
