package cmd

import (
	"github.com/herlon214/ipfs-blockchain/cmd/wallets"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "blockchain",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	RootCmd.AddCommand(wallets.WalletsCmd)
}
