package wallets

import "github.com/spf13/cobra"

var (
	walletFile string
)

var WalletsCmd = &cobra.Command{
	Use:   "wallets",
	Short: "Deal with wallets",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	WalletsCmd.PersistentFlags().StringVarP(&walletFile, "file", "f", "", "wallet path")

	WalletsCmd.AddCommand(NewWallet)
	WalletsCmd.AddCommand(ListCmd)
	WalletsCmd.AddCommand(NewAddrCmd)
}
