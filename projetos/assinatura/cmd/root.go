package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "assinatura",
	Short: "CLI para criação e validação de assinaturas digitais simuladas",
	Long: `Sistema Runner — CLI multiplataforma do HubSaúde.

Permite criar e validar assinaturas digitais simuladas sem necessidade
de configurar ou instalar o Java manualmente.

Exemplos:
  assinatura sign --content "meu documento"
  assinatura validate --content "meu documento" --signature "MOCKED_SIGNATURE_BASE64_=="`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Flags persistentes globais podem ser adicionadas aqui.
}
