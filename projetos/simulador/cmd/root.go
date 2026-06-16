package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// defaultPort é a porta padrão do Simulador: 8443, a própria porta/`server.base-url`
// do hubsaude-simulador.jar (HTTPS). Mantê-la em 8443 deixa os documentos de
// discovery (que emitem https://localhost:8443) coerentes e ainda evita colisão com
// o assinador.jar (8080).
const defaultPort = 8443

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
