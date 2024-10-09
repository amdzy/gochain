package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var address string

var rootCmd = &cobra.Command{
	Use:     "gochain",
	Short:   "GoChain is a demo blockchain",
	Long:    `GoChain is a demo blockchain built for learning purposes`,
	Version: "0.0.1",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	cobra.EnableCommandSorting = false
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(printChainCmd)

	createBlockChainCmd.Flags().StringVarP(&address, "address", "a", "", "The address to send genesis block reward to")
	cobra.MarkFlagRequired(createBlockChainCmd.Flags(), "address")
	rootCmd.AddCommand(createBlockChainCmd)

	getBalanceCmd.Flags().StringVarP(&address, "address", "a", "", "The address to send genesis block reward to")
	cobra.MarkFlagRequired(getBalanceCmd.Flags(), "address")
	rootCmd.AddCommand(getBalanceCmd)

	sendCmd.Flags().StringVarP(&sendFrom, "from", "f", "", "The address to of the user sending")
	sendCmd.Flags().StringVarP(&sendTo, "to", "t", "", "The address to of the user receiving")
	sendCmd.Flags().IntVarP(&sendAmount, "amount", "a", 0, "The amount to send")
	cobra.MarkFlagRequired(sendCmd.Flags(), "from")
	cobra.MarkFlagRequired(sendCmd.Flags(), "to")
	cobra.MarkFlagRequired(sendCmd.Flags(), "amount")
	rootCmd.AddCommand(sendCmd)

	rootCmd.AddCommand(createWalletCmd)

	rootCmd.AddCommand(listAddressesCmd)

	rootCmd.AddCommand(reIndexUTXoCmd)

	startNodeCmd.Flags().StringVarP(&minerAddress, "miner", "a", "", "Enable mining mode and send reward to ADDRESS")
	rootCmd.AddCommand(startNodeCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
