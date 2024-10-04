package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "gochain",
	Short:   "GoChain is a demo blockchain",
	Long:    `GoChain is a demo blockchain built for learning purposes`,
	Version: "0.0.1",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Got Called")
	},
}

func init() {
	rootCmd.AddCommand(printChainCmd)

	addBlockCmd.Flags().StringVarP(&newBlockData, "data", "d", "", "The new block data")
	cobra.MarkFlagRequired(addBlockCmd.Flags(), "data")
	rootCmd.AddCommand(addBlockCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
