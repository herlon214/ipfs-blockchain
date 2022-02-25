package wallets

import (
	"fmt"

	walletsPkg "github.com/herlon214/ipfs-blockchain/pkg/wallets"
	"github.com/spf13/cobra"
)

var NewAddrCmd = &cobra.Command{
	Use:   "newaddr",
	Short: "Create a new wallet address",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Loading wallet from file", walletFile)
		ws, err := walletsPkg.Load(walletFile)
		if err != nil {
			panic(err)
		}

		defer ws.Save()

		wallet, err := ws.NewWallet()
		if err != nil {
			panic(err)
		}

		fmt.Println("Address created:", string(wallet.Address()))
	},
}
