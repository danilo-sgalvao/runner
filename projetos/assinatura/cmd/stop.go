package cmd

import (
	"fmt"
	"os"

	"github.com/danilo-sgalvao/runner/internal/server"
	"github.com/spf13/cobra"
)

var stopPort int

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Encerra o assinador.jar em execução",
	Long: `Encerra o servidor assinador.jar que está rodando em background.

Lê o PID registrado em ~/.hubsaude/assinador.pid, encerra o processo
e limpa o registro.

Exemplos:
  assinatura stop
  assinatura stop --port 9090`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		info, err := server.ReadProcessInfo()
		if err != nil {
			return fmt.Errorf("servidor não encontrado (nenhum processo registrado): %w", err)
		}

		if stopPort != 0 && info.Port != stopPort {
			return fmt.Errorf("nenhum servidor registrado na porta %d (servidor ativo está na porta %d)", stopPort, info.Port)
		}

		if !server.IsResponding(info.Port) {
			fmt.Printf("Servidor na porta %d não está respondendo (já encerrado?). Limpando registro.\n", info.Port)
			server.ClearProcessInfo()
			return nil
		}

		proc, err := os.FindProcess(info.PID)
		if err != nil {
			server.ClearProcessInfo()
			return fmt.Errorf("processo PID %d não encontrado: %w", info.PID, err)
		}

		if err := proc.Kill(); err != nil {
			return fmt.Errorf("falha ao encerrar processo PID %d: %w", info.PID, err)
		}

		server.ClearProcessInfo()
		fmt.Printf("Servidor encerrado (porta %d, PID %d)\n", info.Port, info.PID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().IntVar(&stopPort, "port", 0, "Porta do servidor a encerrar (padrão: porta registrada no ~/.hubsaude/assinador.pid)")
}
