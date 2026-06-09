package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/danilo-sgalvao/runner/shared/jre"
	"github.com/danilo-sgalvao/runner/shared/process"
	"github.com/danilo-sgalvao/runner/simulador/internal/simjar"
	"github.com/danilo-sgalvao/runner/simulador/internal/simserver"
	"github.com/spf13/cobra"
)

var (
	startPort   int
	startSource string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o Simulador do HubSaúde em background",
	Long: `Inicia o Simulador do HubSaúde (validador FHIR) como processo em background.

Verifica se a porta está livre, obtém o simulador.jar (baixando do repositório
da disciplina quando ausente), garante um Java 21+ disponível e sobe o processo.
O PID e a porta são registrados em ~/.hubsaude/simulador.pid.

O primeiro start pode levar cerca de um minuto: o validador carrega os pacotes
FHIR embutidos antes de aceitar requisições.

Exemplos:
  simulador start
  simulador start --port 9443
  simulador start --source https://exemplo/simulador.jar`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		// Reutiliza instância já em execução na mesma porta.
		if info, err := simserver.ReadProcessInfo(); err == nil && info.Port == startPort {
			if simserver.IsResponding(info.Port) {
				fmt.Printf("Simulador já em execução na porta %d (PID %d)\n", info.Port, info.PID)
				return nil
			}
		}

		// US-03.1: verifica se a porta está disponível antes de iniciar.
		if !simserver.IsPortFree(startPort) {
			return fmt.Errorf("a porta %d já está em uso — encerre o processo existente ou use --port", startPort)
		}

		jarPath, err := simjar.Find(startSource)
		if err != nil {
			return err
		}

		javaPath, err := jre.JavaPath()
		if err != nil {
			return fmt.Errorf("Java não disponível: %w", err)
		}

		javaCmd := exec.Command(javaPath, "-jar", jarPath, fmt.Sprintf("--server.port=%d", startPort))
		javaCmd.Stderr = os.Stderr
		process.Detach(javaCmd)

		if err := javaCmd.Start(); err != nil {
			return fmt.Errorf("falha ao iniciar o simulador: %w", err)
		}

		if err := simserver.WriteProcessInfo(simserver.ProcessInfo{PID: javaCmd.Process.Pid, Port: startPort}); err != nil {
			return fmt.Errorf("simulador iniciado (PID %d) mas falhou ao registrar o processo: %w", javaCmd.Process.Pid, err)
		}

		fmt.Printf("Aguardando o simulador ficar pronto na porta %d (pode levar ~1 min no primeiro start)...\n", startPort)
		if err := simserver.WaitUntilReady(startPort, 90*time.Second); err != nil {
			return fmt.Errorf("o simulador não ficou pronto na porta %d: %w", startPort, err)
		}

		fmt.Printf("Simulador em execução na porta %d (PID %d)\n", startPort, javaCmd.Process.Pid)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVar(&startPort, "port", defaultPort, "Porta do Simulador")
	startCmd.Flags().StringVar(&startSource, "source", "", "URL alternativa para baixar o simulador.jar (sobrepõe o release.json)")
}
