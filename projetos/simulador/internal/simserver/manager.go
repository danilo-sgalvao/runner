// Package simserver gerencia o ciclo de vida do simulador.jar (hubsaude-simulador
// — servidor de autorização SMART on FHIR / OAuth2 com mTLS): registro de
// PID/porta, checagem de readiness/status via GET /api/info e verificação de porta
// livre.
//
// Diferenças em relação ao internal/server do assinatura, ditadas pelo contrato do
// jar externo (hubsaude-simulador, verificado ao vivo em 2026-06-15):
//   - o registro ~/.hubsaude/simulador.pid é gravado pelo PRÓPRIO CLI (WriteProcessInfo),
//     pois o jar externo não escreve esse arquivo;
//   - o serviço é HTTPS com certificado self-signed (keystore p12 embutido) e
//     client-auth: want — GETs de probe passam sem certificado de cliente, mas o
//     cliente Go precisa pular a verificação da cadeia (InsecureSkipVerify);
//   - readiness/status usam GET /api/info (200 = no ar), probe estável entre versões do jar
//     (o Spring Actuator não existia no 0.0.0-SNAPSHOT e passou a existir no 0.1.11; o CLI não
//     depende dele);
//   - existe POST /shutdown (graceful); o comando stop tenta /shutdown primeiro e
//     encerra por PID como fallback.
package simserver

import (
	"crypto/tls"
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

// httpClient fala HTTPS com o simulador. O serviço usa certificado self-signed
// (keystore p12 embutido) e client-auth: want — não exige certificado de cliente
// para GETs. Pulamos a verificação da cadeia porque o objetivo é apenas a gerência
// local de ciclo de vida (probe a localhost), não um canal de dados sensível.
var httpClient = &http.Client{
	Timeout: httpTimeout,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	},
}

func pidFile() string {
	if PidFilePath != "" {
		return PidFilePath
	}
	return config.PidPath()
}

func baseURL(port int, path string) string {
	return fmt.Sprintf("https://%s:%d%s", dialHost, port, path)
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

// IsResponding indica se há um servidor atendendo na porta (qualquer status HTTP).
// Diferente de "no ar": basta completar um round-trip — usado para detectar uma
// instância já em execução antes de subir outra.
func IsResponding(port int) bool {
	resp, err := httpClient.Get(baseURL(port, "/api/info"))
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

// Info reflete o corpo de GET /api/info do simulador.
type Info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Probe consulta GET /api/info; 200 significa simulador no ar.
func Probe(port int) (*Info, error) {
	resp, err := httpClient.Get(baseURL(port, "/api/info"))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status inesperado %d em /api/info", resp.StatusCode)
	}
	var info Info
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("resposta de /api/info inválida: %w", err)
	}
	return &info, nil
}

// RequestShutdown envia POST /shutdown ao simulador (encerramento gracioso).
// Retorna nil se o servidor aceitou (HTTP 200).
func RequestShutdown(port int) error {
	resp, err := httpClient.Post(baseURL(port, "/shutdown"), "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("POST /shutdown retornou %d", resp.StatusCode)
	}
	return nil
}

// WaitUntilDown aguarda o simulador parar de responder em /api/info,
// consultando a cada 500ms até o timeout. Retorna nil quando a porta não
// responde mais (sinal de que o processo encerrou).
func WaitUntilDown(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := baseURL(port, "/api/info")
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		resp, err := httpClient.Get(url)
		if err != nil {
			return nil
		}
		resp.Body.Close()
	}
	return fmt.Errorf("simulador ainda responde após %s", timeout)
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
