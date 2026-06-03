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

		if !validateLocal {
			if info, err := server.ReadProcessInfo(); err == nil && server.IsResponding(info.Port) {
				resp, err := server.Validate(info.Port, validateContent, validateSignature)
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
	validateCmd.Flags().BoolVar(&validateLocal, "local", false, "Força modo local (java -jar) mesmo com servidor ativo")
	validateCmd.MarkFlagRequired("content")
	validateCmd.MarkFlagRequired("signature")
}
