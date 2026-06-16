package cmd

import (
	"fmt"

	"github.com/danilo-sgalvao/runner/simulador/internal/simserver"
	"github.com/spf13/cobra"
)

var statusPort int

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Exibe o status atual do Simulador",
	Long: `Consulta o Simulador e informa se ele está em execução.

O status é obtido via GET /api/info. Se a porta não for informada, usa a porta
registrada em ~/.hubsaude/simulador.pid (ou a porta padrão). Um registro órfão
(sem processo respondendo) é limpo automaticamente.

Exemplos:
  simulador status
  simulador status --port 9443`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		info, _ := simserver.ReadProcessInfo()

		port := statusPort
		if port == 0 {
			if info != nil {
				port = info.Port
			} else {
				port = defaultPort
			}
		}

		app, err := simserver.Probe(port)
		if err != nil {
			fmt.Printf("Simulador não está em execução na porta %d\n", port)
			// Reconcilia registro órfão: havia PID registrado para esta porta,
			// mas não há ninguém respondendo.
			if info != nil && info.Port == port {
				simserver.ClearProcessInfo()
			}
			return nil
		}

		if info != nil && info.Port == port {
			fmt.Printf("Simulador em execução na porta %d (PID %d) — %s %s\n", port, info.PID, app.Name, app.Version)
		} else {
			fmt.Printf("Simulador em execução na porta %d — %s %s\n", port, app.Name, app.Version)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().IntVar(&statusPort, "port", 0, "Porta do Simulador a consultar (padrão: porta registrada ou a padrão)")
}
