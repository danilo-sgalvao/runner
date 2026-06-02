package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestStartCmd_EstaRegistradoNoRoot(t *testing.T) {
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "start" {
			return
		}
	}
	t.Fatal("comando 'start' não está registrado no rootCmd")
}

func TestStartCmd_ContemFlagPort(t *testing.T) {
	flag := startCmd.Flags().Lookup("port")
	if flag == nil {
		t.Fatal("flag --port não encontrada no comando start")
	}
	if flag.DefValue != "8080" {
		t.Errorf("valor padrão de --port deveria ser 8080, obteve: %s", flag.DefValue)
	}
}

func TestStartCmd_ContemFlagTimeout(t *testing.T) {
	flag := startCmd.Flags().Lookup("timeout")
	if flag == nil {
		t.Fatal("flag --timeout não encontrada no comando start")
	}
	if flag.DefValue != "0" {
		t.Errorf("valor padrão de --timeout deveria ser 0, obteve: %s", flag.DefValue)
	}
}

func TestStartCmd_NomeCorreto(t *testing.T) {
	if startCmd.Name() != "start" {
		t.Errorf("nome do comando deveria ser 'start', obteve: %s", startCmd.Name())
	}
}

func TestStartCmd_ShortNaoVazio(t *testing.T) {
	if startCmd.Short == "" {
		t.Error("descrição curta do comando start não deveria ser vazia")
	}
}

func TestStartCmd_PortNaoEObrigatoria(t *testing.T) {
	flag := startCmd.Flags().Lookup("port")
	if flag == nil {
		t.Fatal("flag --port não encontrada")
	}
	_, obrigatorio := flag.Annotations[cobra.BashCompOneRequiredFlag]
	if obrigatorio {
		t.Error("flag --port não deveria ser obrigatória (tem valor padrão)")
	}
}
