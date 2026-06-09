package simserver

import (
	"fmt"
	"net/http"
	"time"
)

// WaitUntilReady aguarda o simulador ficar PRONTO na porta indicada, consultando
// GET /actuator/health/readiness a cada 500ms até esgotar o timeout. O endpoint
// responde 200 quando readinessState=UP e 503 durante o cold start (carga dos
// pacotes FHIR), então este polling atravessa o warm-up até o serviço aceitar uso.
//
// O timeout deve ser generoso (≥60s): o cold start do validador é de ~20s mais o
// warm-up lazy do HAPI na primeira validação.
func WaitUntilReady(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: time.Second}
	url := actuatorURL(port, "/actuator/health/readiness")
	for time.Now().Before(deadline) {
		time.Sleep(500 * time.Millisecond)
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
	return fmt.Errorf("simulador não ficou pronto (readiness) após %s", timeout)
}
