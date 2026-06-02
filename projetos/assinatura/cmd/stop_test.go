package cmd

import (
	"testing"
)

func TestStopCmd_EstaRegistradoNoRoot(t *testing.T) {
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "stop" {
			return
		}
	}
	t.Fatal("comando 'stop' não está registrado no rootCmd")
}

func TestStopCmd_ContemFlagPort(t *testing.T) {
	flag := stopCmd.Flags().Lookup("port")
	if flag == nil {
		t.Fatal("flag --port não encontrada no comando stop")
	}
}

func TestStopCmd_NomeCorreto(t *testing.T) {
	if stopCmd.Name() != "stop" {
		t.Errorf("nome do comando deveria ser 'stop', obteve: %s", stopCmd.Name())
	}
}

func TestStopCmd_ShortNaoVazio(t *testing.T) {
	if stopCmd.Short == "" {
		t.Error("descrição curta do comando stop não deveria ser vazia")
	}
}

func TestStopCmd_PortPadraoZero(t *testing.T) {
	flag := stopCmd.Flags().Lookup("port")
	if flag == nil {
		t.Fatal("flag --port não encontrada")
	}
	if flag.DefValue != "0" {
		t.Errorf("valor padrão de --port deveria ser 0 (porta do registro), obteve: %s", flag.DefValue)
	}
}
