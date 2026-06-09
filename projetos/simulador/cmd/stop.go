package cmd

import (
	"fmt"
	"os"

	"github.com/danilo-sgalvao/runner/simulador/internal/simserver"
	"github.com/spf13/cobra"
)

var stopPort int

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Encerra o Simulador em execução",
	Long: `Encerra o Simulador do HubSaúde que está rodando em background.

Lê o PID registrado em ~/.hubsaude/simulador.pid e encerra o processo. O
validador não expõe um endpoint de shutdown, então o encerramento é feito pelo
PID registrado.

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
