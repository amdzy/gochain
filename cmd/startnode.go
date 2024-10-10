package cmd

import (
	"amdzy/gochain/pkg/server"
	"amdzy/gochain/pkg/wallet"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func NewStartNodeCommand() *cobra.Command {

	var minerAddress string
	var startNodeCmd = &cobra.Command{
		Use:   "startnode",
		Short: "Start Node",
		Long:  "Start Node",
		Run: func(cmd *cobra.Command, args []string) {
			if len(minerAddress) > 0 {
				validAddr := wallet.ValidateAddress(minerAddress)
				if !validAddr {
					log.Fatal("Wrong miner address!")
				} else {
					fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
				}
			}

			err := server.StartServer("6000", minerAddress)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	startNodeCmd.Flags().StringVarP(&minerAddress, "miner", "a", "", "Enable mining mode and send reward to ADDRESS")

	return startNodeCmd
}
