package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/danilo-sgalvao/runner/assinatura/internal/jar"
	"github.com/danilo-sgalvao/runner/assinatura/internal/server"
	"github.com/danilo-sgalvao/runner/shared/jre"
	"github.com/danilo-sgalvao/runner/shared/process"
	"github.com/spf13/cobra"
)

var startPort    int
var startTimeout int

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o assinador.jar no modo servidor",
	Long: `Inicia o assinador.jar como servidor HTTP em background.

O servidor fica disponível para requisições sign e validate com menor latência,
eliminando o overhead de cold start do Java a cada operação.

O PID e a porta são registrados em ~/.hubsaude/assinador.pid para que os
comandos sign, validate e stop possam gerenciar o ciclo de vida do servidor.

Exemplos:
  assinatura start
  assinatura start --port 9090
  assinatura start --port 9090 --timeout 30`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		if info, err := server.ReadProcessInfo(); err == nil && info.Port == startPort {
			if server.IsResponding(info.Port) {
				fmt.Printf("Servidor já em execução na porta %d (PID %d)\n", info.Port, info.PID)
				return nil
			}
		}

		jarPath, err := jar.Find()
		if err != nil {
			return err
		}

		javaPath, err := jre.JavaPath()
		if err != nil {
			return fmt.Errorf("Java não disponível: %w", err)
		}

		javaArgs := []string{"-jar", jarPath, "serve", "--port", fmt.Sprintf("%d", startPort)}
		javaCmd := exec.Command(javaPath, javaArgs...)

		env := os.Environ()
		if startTimeout > 0 {
			env = append(env, fmt.Sprintf("HUBSAUDE_TIMEOUT_MINUTES=%d", startTimeout))
		}
		javaCmd.Env = env
		javaCmd.Stderr = os.Stderr

		process.Detach(javaCmd)

		if err := javaCmd.Start(); err != nil {
			return fmt.Errorf("falha ao iniciar servidor: %w", err)
		}

		fmt.Printf("Aguardando servidor iniciar na porta %d...\n", startPort)
		if err := server.WaitUntilReady(startPort, 30*time.Second); err != nil {
			return fmt.Errorf("servidor não respondeu após 30 segundos — verifique se a porta %d está disponível", startPort)
		}

		fmt.Printf("Servidor iniciado na porta %d (PID %d)\n", startPort, javaCmd.Process.Pid)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVar(&startPort, "port", 8080, "Porta do servidor")
	startCmd.Flags().IntVar(&startTimeout, "timeout", 0, "Inatividade máxima em minutos antes de encerrar automaticamente (0 = desativado)")
}
