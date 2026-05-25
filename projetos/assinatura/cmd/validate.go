package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/danilo-sgalvao/runner/internal/jre"
	"github.com/spf13/cobra"
)

var validateContent string
var validateSignature string

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Valida uma assinatura digital simulada",
	Long: `Invoca o assinador.jar para verificar se uma assinatura digital é válida.

O Java é detectado automaticamente. Se não estiver instalado, será baixado
e configurado em ~/.hubsaude/jre sem necessidade de interação do usuário.

Exemplos:
  assinatura validate --content "documento" --signature "MOCKED_SIGNATURE_BASE64_=="`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		jarPath, err := encontrarJar()
		if err != nil {
			return err
		}

		javaPath, err := jre.JavaPath()
		if err != nil {
			return fmt.Errorf("Java não disponível: %w", err)
		}

		javaCmd := exec.Command(javaPath, "-jar", jarPath,
			"validate",
			"--content", validateContent,
			"--signature", validateSignature,
		)
		javaCmd.Stdout = os.Stdout
		javaCmd.Stderr = os.Stderr

		return javaCmd.Run()
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().StringVar(&validateContent, "content", "", "Conteúdo original (obrigatório)")
	validateCmd.Flags().StringVar(&validateSignature, "signature", "", "Assinatura a ser validada (obrigatório)")
	validateCmd.MarkFlagRequired("content")
	validateCmd.MarkFlagRequired("signature")
}
