package cmd

import (
	"github.com/spf13/cobra"
)

func NewDefaultCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:     "gochain",
		Short:   "GoChain is a demo blockchain",
		Long:    `GoChain is a demo blockchain built for learning purposes`,
		Version: "0.0.1",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cobra.EnableCommandSorting = false
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(NewPrintChainCommand())
	rootCmd.AddCommand(NewCreateBlockchainCommand())
	rootCmd.AddCommand(NewGetBalanceCommand())
	rootCmd.AddCommand(NewSendCmdCommand())
	rootCmd.AddCommand(NewCreateWalletCommand())
	rootCmd.AddCommand(NewListAddressesCommand())
	rootCmd.AddCommand(NewReIndexUTXoCommand())
	rootCmd.AddCommand(NewStartNodeCommand())

	return rootCmd
}
