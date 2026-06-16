package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version é injetada em tempo de build via -ldflags; "dev" em builds locais.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Exibe a versão do CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("simulador " + Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
