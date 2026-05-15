package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestSignCmd_EstaRegistradoNoRoot(t *testing.T) {
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "sign" {
			return
		}
	}
	t.Fatal("comando 'sign' não está registrado no rootCmd")
}

func TestSignCmd_ContemFlagContent(t *testing.T) {
	flag := signCmd.Flags().Lookup("content")
	if flag == nil {
		t.Fatal("flag --content não encontrada no comando sign")
	}
}

func TestSignCmd_ContentEObrigatorio(t *testing.T) {
	flag := signCmd.Flags().Lookup("content")
	if flag == nil {
		t.Fatal("flag --content não encontrada")
	}
	annotations := flag.Annotations
	if _, ok := annotations[cobra.BashCompOneRequiredFlag]; !ok {
		t.Error("flag --content deveria ser marcada como obrigatória com MarkFlagRequired")
	}
}

func TestSignCmd_NaoContemFlagAlgorithm(t *testing.T) {
	flag := signCmd.Flags().Lookup("algorithm")
	if flag != nil {
		t.Fatal("flag --algorithm foi removida e não deveria existir no comando sign")
	}
}

func TestSignCmd_NomeCorreto(t *testing.T) {
	if signCmd.Name() != "sign" {
		t.Errorf("nome do comando deveria ser 'sign', obteve: %s", signCmd.Name())
	}
}

func TestSignCmd_ShortNaoVazio(t *testing.T) {
	if signCmd.Short == "" {
		t.Error("descrição curta do comando sign não deveria ser vazia")
	}
}
