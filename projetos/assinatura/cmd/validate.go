package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var validateContent string
var validateSignature string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Valida uma assinatura digital simulada",
	Run: func(cmd *cobra.Command, args []string) {
		if validateContent == "" {
			fmt.Println("Erro: --content é obrigatório.")
			os.Exit(1)
		}

		if validateSignature == "" {
			fmt.Println("Erro: --signature é obrigatório.")
			os.Exit(1)
		}

		jarPath := encontrarJar()

		javaCmd := exec.Command("java", "-jar", jarPath,
			"validate",
			"--content", validateContent,
			"--signature", validateSignature,
		)

		javaCmd.Stdout = os.Stdout
		javaCmd.Stderr = os.Stderr

		err := javaCmd.Run()
		if err != nil {
			fmt.Println("Erro ao executar assinador.jar:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVar(&validateContent, "content", "", "Conteúdo original (obrigatório)")
	validateCmd.Flags().StringVar(&validateSignature, "signature", "", "Assinatura a ser validada (obrigatório)")
}