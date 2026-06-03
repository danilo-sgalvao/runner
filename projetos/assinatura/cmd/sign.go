package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/danilo-sgalvao/runner/internal/jar"
	"github.com/danilo-sgalvao/runner/internal/jre"
	"github.com/danilo-sgalvao/runner/internal/server"
	"github.com/spf13/cobra"
)

var signContent string
var signLocal   bool

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Cria uma assinatura digital simulada",
	Long: `Invoca o assinador.jar para criar uma assinatura digital simulada.

Por padrão, usa o servidor HTTP em execução quando disponível (menor latência).
Se não houver servidor ativo ou --local for especificado, invoca o
assinador.jar diretamente via java -jar.

Exemplos:
  assinatura sign --content "documento"
  assinatura sign --content "documento" --local`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		if !signLocal {
			if info, err := server.ReadProcessInfo(); err == nil && server.IsResponding(info.Port) {
				resp, err := server.Sign(info.Port, signContent, "")
				if err != nil {
					return fmt.Errorf("erro ao chamar servidor: %w", err)
				}
				fmt.Printf("Assinatura: %s\nVálido: %v\nMensagem: %s\n",
					resp.Signature, resp.Valid, resp.Message)
				return nil
			}
		}

		// Modo local: invoca java -jar diretamente
		jarPath, err := jar.Find()
		if err != nil {
			return err
		}

		javaPath, err := jre.JavaPath()
		if err != nil {
			return fmt.Errorf("Java não disponível: %w", err)
		}

		javaCmd := exec.Command(javaPath, "-jar", jarPath, "sign", "--content", signContent)
		javaCmd.Stdout = os.Stdout
		javaCmd.Stderr = os.Stderr
		return javaCmd.Run()
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.Flags().StringVar(&signContent, "content", "", "Conteúdo a ser assinado (obrigatório)")
	signCmd.Flags().BoolVar(&signLocal, "local", false, "Força modo local (java -jar) mesmo com servidor ativo")
	signCmd.MarkFlagRequired("content")
}
