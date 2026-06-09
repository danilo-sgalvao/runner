package server_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/danilo-sgalvao/runner/assinatura/internal/server"
)

func TestReadProcessInfo_arquivoValido(t *testing.T) {
	dir := t.TempDir()
	server.PidFilePath = filepath.Join(dir, "assinador.pid")
	defer func() { server.PidFilePath = "" }()

	info := server.ProcessInfo{PID: 1234, Port: 8080}
	data, _ := json.Marshal(info)
	if err := os.WriteFile(server.PidFilePath, data, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := server.ReadProcessInfo()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if got.PID != 1234 || got.Port != 8080 {
		t.Errorf("esperado PID=1234 Port=8080, obteve PID=%d Port=%d", got.PID, got.Port)
	}
}

func TestReadProcessInfo_arquivoInexistente(t *testing.T) {
	server.PidFilePath = filepath.Join(t.TempDir(), "nao-existe.pid")
	defer func() { server.PidFilePath = "" }()

	_, err := server.ReadProcessInfo()
	if err == nil {
		t.Fatal("esperava erro para arquivo inexistente")
	}
}

func TestReadProcessInfo_jsonCorrompido(t *testing.T) {
	dir := t.TempDir()
	server.PidFilePath = filepath.Join(dir, "assinador.pid")
	defer func() { server.PidFilePath = "" }()

	if err := os.WriteFile(server.PidFilePath, []byte("não é json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := server.ReadProcessInfo()
	if err == nil {
		t.Fatal("esperava erro para JSON corrompido")
	}
}

func TestClearProcessInfo_removeArquivo(t *testing.T) {
	dir := t.TempDir()
	server.PidFilePath = filepath.Join(dir, "assinador.pid")
	defer func() { server.PidFilePath = "" }()

	if err := os.WriteFile(server.PidFilePath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	server.ClearProcessInfo()

	if _, err := os.Stat(server.PidFilePath); !os.IsNotExist(err) {
		t.Error("arquivo deveria ter sido removido por ClearProcessInfo")
	}
}
