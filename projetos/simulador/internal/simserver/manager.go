// Package simserver gerencia o ciclo de vida do simulador.jar (validador FHIR
// hubsaude-validador-api): registro de PID/porta, checagem de readiness/health via
// Spring Actuator e verificação de porta livre.
//
// Diferenças em relação ao internal/server do assinatura, ditadas pelo contrato do
// jar externo (confirmado na v0.1.10):
//   - o registro ~/.hubsaude/simulador.pid é gravado pelo PRÓPRIO CLI (WriteProcessInfo),
//     pois o jar externo não escreve esse arquivo;
//   - readiness/health usam os endpoints do Actuator (/actuator/health[...]), não /health;
//   - não há endpoint de shutdown: o encerramento é por PID (no comando stop).
package simserver

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/danilo-sgalvao/runner/simulador/internal/config"
)

// ProcessInfo descreve a instância do simulador.jar registrada pelo CLI.
type ProcessInfo struct {
	PID  int `json:"pid"`
	Port int `json:"port"`
}

// PidFilePath é o caminho do arquivo de registro; pode ser sobrescrito em testes.
var PidFilePath = ""

// dialHost é o host usado nas chamadas HTTP locais; sobrescrito em testes para
// casar com o endereço do httptest (127.0.0.1).
var dialHost = "localhost"

const httpTimeout = 2 * time.Second

func pidFile() string {
	if PidFilePath != "" {
		return PidFilePath
	}
	return config.PidPath()
}

func actuatorURL(port int, path string) string {
	return fmt.Sprintf("http://%s:%d%s", dialHost, port, path)
}

// WriteProcessInfo grava o registro {pid, port} em ~/.hubsaude/simulador.pid.
// É o CLI que registra (o jar externo não o faz).
func WriteProcessInfo(info ProcessInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return os.WriteFile(pidFile(), data, 0644)
}

// ReadProcessInfo lê o registro gravado por WriteProcessInfo.
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

// ClearProcessInfo remove o registro do simulador.
func ClearProcessInfo() {
	_ = os.Remove(pidFile())
}

// IsResponding indica se há um servidor HTTP atendendo na porta (qualquer status).
// Diferente do UP/DOWN: durante o cold start o /actuator/health responde 503, mas o
// processo já está vivo — para detectar instância existente, basta responder a HTTP.
func IsResponding(port int) bool {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(actuatorURL(port, "/actuator/health"))
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

// HealthStatus reflete o agregado de /actuator/health (campo "status": UP/DOWN/...).
type HealthStatus struct {
	Status string `json:"status"`
}

// Health consulta /actuator/health e devolve o status agregado do simulador.
func Health(port int) (*HealthStatus, error) {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(actuatorURL(port, "/actuator/health"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var h HealthStatus
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		return nil, fmt.Errorf("resposta de health inválida: %w", err)
	}
	return &h, nil
}

// IsPortFree informa se a porta TCP pode ser vinculada (livre para iniciar o
// simulador). Usado por start antes de subir o processo (US-03.1).
func IsPortFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}
