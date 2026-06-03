package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/danilo-sgalvao/runner/internal/config"
)

// ProcessInfo descreve a instância do assinador.jar em execução.
// O servidor grava este arquivo em ~/.hubsaude/assinador.pid ao iniciar.
type ProcessInfo struct {
	PID  int `json:"pid"`
	Port int `json:"port"`
}

// PidFilePath é o caminho do arquivo de registro; pode ser sobrescrito em testes.
var PidFilePath = ""

func pidFile() string {
	if PidFilePath != "" {
		return PidFilePath
	}
	return config.PidPath()
}

// ReadProcessInfo lê o registro salvo pelo servidor na inicialização.
func ReadProcessInfo() (*ProcessInfo, error) {
	data, err := os.ReadFile(pidFile())
	if err != nil {
		return nil, err
	}
	var info ProcessInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("arquivo de PID corrompido: %w", err)
	}
	return &info, nil
}

// IsResponding verifica se o servidor na porta indicada está respondendo ao /health.
func IsResponding(port int) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", port))
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ClearProcessInfo remove o arquivo de registro do servidor.
func ClearProcessInfo() {
	_ = os.Remove(pidFile())
}
