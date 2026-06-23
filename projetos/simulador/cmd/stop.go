package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/danilo-sgalvao/runner/simulador/internal/simserver"
	"github.com/spf13/cobra"
)

var stopPort int

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Encerra o Simulador em execução",
	Long: `Encerra o Simulador do HubSaúde que está rodando em background.

Tenta encerramento gracioso via POST /shutdown (contrato do jar); se o servidor
não responder ou não parar em 10s, encerra forçosamente pelo PID registrado em
~/.hubsaude/simulador.pid.

Exemplos:
  simulador stop
  simulador stop --port 9443`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		info, err := simserver.ReadProcessInfo()
		if err != nil {
			return fmt.Errorf("simulador não encontrado (nenhum processo registrado): %w", err)
		}

		if stopPort != 0 && info.Port != stopPort {
			return fmt.Errorf("nenhum simulador registrado na porta %d (registro ativo está na porta %d)", stopPort, info.Port)
		}

		if err := simserver.RequestShutdown(info.Port); err == nil {
			if simserver.WaitUntilDown(info.Port, 10*time.Second) == nil {
				simserver.ClearProcessInfo()
				fmt.Printf("Simulador encerrado via /shutdown (porta %d, PID %d)\n", info.Port, info.PID)
				return nil
			}
		}

		proc, err := os.FindProcess(info.PID)
		if err != nil {
			simserver.ClearProcessInfo()
			return fmt.Errorf("processo PID %d não encontrado: %w", info.PID, err)
		}

		if err := proc.Kill(); err != nil {
			return fmt.Errorf("falha ao encerrar o processo PID %d: %w", info.PID, err)
		}

		simserver.ClearProcessInfo()
		fmt.Printf("Simulador encerrado (porta %d, PID %d)\n", info.Port, info.PID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().IntVar(&stopPort, "port", 0, "Porta do Simulador a encerrar (padrão: porta registrada em ~/.hubsaude/simulador.pid)")
}
