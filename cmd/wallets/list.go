package wallets

import (
	"fmt"

	walletsPkg "github.com/herlon214/ipfs-blockchain/pkg/wallets"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "ls",
	Short: "List wallets addresses",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Loading wallet from file", walletFile)
		ws, err := walletsPkg.Load(walletFile)
		if err != nil {
			panic(err)
		}

		fmt.Println("Listing", len(ws.Items), "addresses:")
		for _, wallet := range ws.Items {
			fmt.Println("-->", string(wallet.Address()))
		}
	},
}
