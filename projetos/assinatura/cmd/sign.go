package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/danilo-sgalvao/runner/internal/jre"
	"github.com/spf13/cobra"
)

var signContent string
var signAlgorithm string

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Cria uma assinatura digital simulada",
	Long: `Invoca o assinador.jar para criar uma assinatura digital simulada.

O Java é detectado automaticamente. Se não estiver instalado, será baixado
e configurado em ~/.hubsaude/jre sem necessidade de interação do usuário.

Exemplos:
  assinatura sign --content "documento"
  assinatura sign --content "documento" --algorithm SHA512withRSA`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		jarPath, err := encontrarJar()
		if err != nil {
			return err
		}

		javaPath, err := jre.JavaPath()
		if err != nil {
			return fmt.Errorf("Java não disponível: %w", err)
		}

		javaCmd := exec.Command(javaPath, "-jar", jarPath,
			"sign",
			"--content", signContent,
			"--algorithm", signAlgorithm,
		)
		javaCmd.Stdout = os.Stdout
		javaCmd.Stderr = os.Stderr

		return javaCmd.Run()
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.Flags().StringVar(&signContent, "content", "", "Conteúdo a ser assinado (obrigatório)")
	signCmd.Flags().StringVar(&signAlgorithm, "algorithm", "SHA256withRSA",
		"Algoritmo de assinatura: SHA256withRSA (padrão) ou SHA512withRSA")
	signCmd.MarkFlagRequired("content")
}
