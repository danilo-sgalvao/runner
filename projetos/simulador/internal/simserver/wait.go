package simserver

import (
	"fmt"
	"net/http"
	"time"
)

// WaitUntilReady aguarda o simulador ficar PRONTO na porta indicada, consultando
// GET /api/info a cada 500ms até responder 200 ou esgotar o timeout. Reusa o
// httpClient (HTTPS + InsecureSkipVerify) do pacote.
//
// O serviço sobe rápido (~3s) — não há cold start pesado de pacotes FHIR. Ainda
// assim o timeout deve dar folga no primeiro start, que pode baixar/preparar o JRE
// antes de subir o processo.
func WaitUntilReady(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := baseURL(port, "/api/info")
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		resp, err := httpClient.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
	return fmt.Errorf("simulador não ficou pronto (/api/info) após %s", timeout)
}
