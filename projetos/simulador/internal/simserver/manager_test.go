package simserver

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

// serverPort sobe um httptest HTTPS com o handler dado, aponta dialHost para
// 127.0.0.1 (endereço do httptest) e devolve a porta em que ele escuta. Usa
// NewTLSServer porque o simulador real serve HTTPS; o httpClient do pacote pula a
// verificação do certificado (InsecureSkipVerify), aceitando o cert do httptest.
func serverPort(t *testing.T, h http.Handler) int {
	t.Helper()
	srv := httptest.NewTLSServer(h)
	t.Cleanup(srv.Close)

	orig := dialHost
	dialHost = "127.0.0.1"
	t.Cleanup(func() { dialHost = orig })

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("URL do httptest inválida: %v", err)
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatalf("porta do httptest inválida: %v", err)
	}
	return port
}

func TestProcessInfo_RoundTrip(t *testing.T) {
	PidFilePath = filepath.Join(t.TempDir(), "simulador.pid")
	t.Cleanup(func() { PidFilePath = "" })

	if err := WriteProcessInfo(ProcessInfo{PID: 4321, Port: 8443}); err != nil {
		t.Fatalf("WriteProcessInfo: %v", err)
	}
	info, err := ReadProcessInfo()
	if err != nil {
		t.Fatalf("ReadProcessInfo: %v", err)
	}
	if info.PID != 4321 || info.Port != 8443 {
		t.Errorf("registro = %+v, esperava {4321 8443}", info)
	}

	ClearProcessInfo()
	if _, err := ReadProcessInfo(); err == nil {
		t.Error("esperava erro ao ler registro após ClearProcessInfo")
	}
}

func TestIsResponding(t *testing.T) {
	// Qualquer round-trip conta como "respondendo" — mesmo um status != 200
	// (ex.: GET /shutdown no jar real responde 500: método não suportado).
	port := serverPort(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	if !IsResponding(port) {
		t.Error("esperava IsResponding=true para servidor que responde (ainda que 500)")
	}

	if IsResponding(freePort(t)) {
		t.Error("esperava IsResponding=false quando nada escuta na porta")
	}
}

func TestProbe(t *testing.T) {
	port := serverPort(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/info" {
			t.Errorf("caminho inesperado: %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"name":"HubSaúde Simulador","version":"0.0.0-SNAPSHOT"}`)
	}))

	info, err := Probe(port)
	if err != nil {
		t.Fatalf("Probe: %v", err)
	}
	if info.Name != "HubSaúde Simulador" || info.Version != "0.0.0-SNAPSHOT" {
		t.Errorf("info = %+v, esperava name/version preenchidos", info)
	}
}

func TestWaitUntilReady_FicaPronto(t *testing.T) {
	var hits atomic.Int32
	// Responde 503 nas primeiras consultas e 200 a partir da terceira.
	port := serverPort(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) >= 3 {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))

	if err := WaitUntilReady(port, 5*time.Second); err != nil {
		t.Fatalf("esperava ficar pronto, obteve: %v", err)
	}
}

func TestWaitUntilReady_Timeout(t *testing.T) {
	port := serverPort(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))

	if err := WaitUntilReady(port, 1200*time.Millisecond); err == nil {
		t.Fatal("esperava timeout quando readiness nunca fica UP")
	}
}

func TestIsPortFree(t *testing.T) {
	if !IsPortFree(freePort(t)) {
		t.Error("esperava IsPortFree=true para porta livre")
	}

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("não foi possível ocupar uma porta: %v", err)
	}
	defer ln.Close()
	busy := ln.Addr().(*net.TCPAddr).Port
	if IsPortFree(busy) {
		t.Errorf("esperava IsPortFree=false para porta ocupada %d", busy)
	}
}

// freePort obtém uma porta provavelmente livre (abre e fecha um listener).
func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("não foi possível obter porta livre: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}
