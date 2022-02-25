package wallets

import (
	"fmt"

	walletsPkg "github.com/herlon214/ipfs-blockchain/pkg/wallets"
	"github.com/spf13/cobra"
)

var NewWallet = &cobra.Command{
	Use:   "new",
	Short: "Create a new wallet",
	Run: func(cmd *cobra.Command, args []string) {
		ws, err := walletsPkg.New(walletFile)
		if err != nil {
			panic(err)
		}

		defer ws.Save()

		fmt.Println("Wallet created!")

		wallet, err := ws.NewWallet()
		if err != nil {
			panic(err)
		}

		fmt.Println("Initial address created:", string(wallet.Address()))

	},
}
