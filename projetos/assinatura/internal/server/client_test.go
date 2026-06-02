package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danilo-sgalvao/runner/internal/server"
)

func TestSign_sucesso(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sign" || r.Method != http.MethodPost {
			t.Errorf("caminho/método inesperado: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(server.SignatureResponse{
			Signature: "SIG123",
			Valid:      true,
			Message:    "ok",
		})
	}))
	defer ts.Close()

	port := httpserverPort(t, ts)
	resp, err := server.Sign(port, "documento", "")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resp.Signature != "SIG123" || !resp.Valid {
		t.Errorf("resposta inesperada: %+v", resp)
	}
}

func TestValidate_sucesso(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/validate" {
			t.Errorf("caminho inesperado: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(server.SignatureResponse{
			Signature: "SIG123",
			Valid:      true,
			Message:    "válida",
		})
	}))
	defer ts.Close()

	port := httpserverPort(t, ts)
	resp, err := server.Validate(port, "documento", "SIG123")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !resp.Valid {
		t.Errorf("esperava Valid=true, obteve %v", resp.Valid)
	}
}

func TestSign_servidorForaDo(t *testing.T) {
	_, err := server.Sign(19999, "documento", "")
	if err == nil {
		t.Fatal("esperava erro para servidor indisponível")
	}
}

// httpserverPort extrai a porta numérica de um httptest.Server.
func httpserverPort(t *testing.T, ts *httptest.Server) int {
	t.Helper()
	addr := strings.TrimPrefix(ts.URL, "http://")
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		t.Fatalf("URL inesperada: %s", ts.URL)
	}
	var port int
	if _, err := fmt.Sscanf(parts[1], "%d", &port); err != nil {
		t.Fatalf("porta inválida em %s: %v", ts.URL, err)
	}
	return port
}
