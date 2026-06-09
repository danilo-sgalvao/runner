package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// defaultPort é a porta padrão do Simulador. Diferente do assinador.jar (8080),
// usa 8081 para que ambos possam coexistir na mesma máquina sem colisão.
const defaultPort = 8081

var rootCmd = &cobra.Command{
	Use:   "simulador",
	Short: "CLI para gerenciar o ciclo de vida do Simulador do HubSaúde",
	Long: `Sistema Runner — CLI multiplataforma do HubSaúde.

Inicia, encerra e monitora o Simulador do HubSaúde (validador FHIR) sem
necessidade de configurar ou instalar o Java manualmente. O simulador.jar é
obtido automaticamente do repositório da disciplina quando ausente.

Exemplos:
  simulador start
  simulador status
  simulador stop`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
