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
//   - readiness/status usam GET /api/info (200 = no ar); não há Spring Actuator
//     (/actuator/** responde 500, não 404);
//   - existe POST /shutdown (graceful), mas o comando stop encerra por PID.
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
