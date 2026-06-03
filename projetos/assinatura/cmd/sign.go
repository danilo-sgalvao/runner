package cmd

import (
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

		handled, err := runViaServer(signLocal, func(port int) (*server.SignatureResponse, error) {
			return server.Sign(port, signContent, "")
		})
		if handled {
			return err
		}

		return runViaJar("sign", "--content", signContent)
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.Flags().StringVar(&signContent, "content", "", "Conteúdo a ser assinado (obrigatório)")
	signCmd.Flags().BoolVar(&signLocal, "local", false, "Força modo local (java -jar) mesmo com servidor ativo")
	signCmd.MarkFlagRequired("content")
}
