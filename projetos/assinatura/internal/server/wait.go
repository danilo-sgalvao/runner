package server

import (
	"fmt"
	"net/http"
	"time"
)

// WaitUntilReady aguarda o servidor responder ao /health na porta indicada,
// consultando a cada 500ms até esgotar o timeout. Retorna nil assim que o
// servidor responde HTTP 200; caso contrário, retorna erro ao expirar.
func WaitUntilReady(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: time.Second}
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", port))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
	return fmt.Errorf("servidor não respondeu após %s", timeout)
}
