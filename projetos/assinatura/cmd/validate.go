package cmd

import (
	"github.com/danilo-sgalvao/runner/assinatura/internal/server"
	"github.com/spf13/cobra"
)

var validateContent   string
var validateSignature string
var validateLocal     bool

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Valida uma assinatura digital simulada",
	Long: `Invoca o assinador.jar para verificar se uma assinatura digital é válida.

Por padrão, usa o servidor HTTP em execução quando disponível (menor latência).
Se não houver servidor ativo ou --local for especificado, invoca o
assinador.jar diretamente via java -jar.

Exemplos:
  assinatura validate --content "documento" --signature "MOCKED_SIGNATURE_BASE64_=="
  assinatura validate --content "documento" --signature "MOCKED..." --local`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		handled, err := runViaServer(validateLocal, func(port int) (*server.SignatureResponse, error) {
			return server.Validate(port, validateContent, validateSignature)
		})
		if handled {
			return err
		}

		return runViaJar("validate",
			"--content", validateContent,
			"--signature", validateSignature,
		)
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVar(&validateContent, "content", "", "Conteúdo original (obrigatório)")
	validateCmd.Flags().StringVar(&validateSignature, "signature", "", "Assinatura a ser validada (obrigatório)")
	validateCmd.Flags().BoolVar(&validateLocal, "local", false, "Força modo local (java -jar) mesmo com servidor ativo")
	validateCmd.MarkFlagRequired("content")
	validateCmd.MarkFlagRequired("signature")
}
