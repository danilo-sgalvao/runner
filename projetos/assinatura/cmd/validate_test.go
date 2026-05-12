package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateCmd_EstaRegistradoNoRoot(t *testing.T) {
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "validate" {
			return
		}
	}
	t.Fatal("comando 'validate' não está registrado no rootCmd")
}

func TestValidateCmd_ContemFlagContent(t *testing.T) {
	flag := validateCmd.Flags().Lookup("content")
	if flag == nil {
		t.Fatal("flag --content não encontrada no comando validate")
	}
}

func TestValidateCmd_ContentEObrigatorio(t *testing.T) {
	flag := validateCmd.Flags().Lookup("content")
	if flag == nil {
		t.Fatal("flag --content não encontrada")
	}
	annotations := flag.Annotations
	if _, ok := annotations[cobra.BashCompOneRequiredFlag]; !ok {
		t.Error("flag --content deveria ser marcada como obrigatória com MarkFlagRequired")
	}
}

func TestValidateCmd_ContemFlagSignature(t *testing.T) {
	flag := validateCmd.Flags().Lookup("signature")
	if flag == nil {
		t.Fatal("flag --signature não encontrada no comando validate")
	}
}

func TestValidateCmd_SignatureEObrigatorio(t *testing.T) {
	flag := validateCmd.Flags().Lookup("signature")
	if flag == nil {
		t.Fatal("flag --signature não encontrada")
	}
	annotations := flag.Annotations
	if _, ok := annotations[cobra.BashCompOneRequiredFlag]; !ok {
		t.Error("flag --signature deveria ser marcada como obrigatória com MarkFlagRequired")
	}
}

func TestValidateCmd_NomeCorreto(t *testing.T) {
	if validateCmd.Name() != "validate" {
		t.Errorf("nome do comando deveria ser 'validate', obteve: %s", validateCmd.Name())
	}
}

func TestValidateCmd_ShortNaoVazio(t *testing.T) {
	if validateCmd.Short == "" {
		t.Error("descrição curta do comando validate não deveria ser vazia")
	}
}
