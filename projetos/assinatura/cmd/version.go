package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versão do CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("assinatura v0.1.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}