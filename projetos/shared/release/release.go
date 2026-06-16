// Package release centraliza a leitura do metadado release.json (informações do
// JRE por plataforma e a URL dos jars), servindo de fonte única compartilhada
// pelos CLIs assinatura e simulador (pacotes jar/simjar) e pelo provisionamento
// do JRE.
package release

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/danilo-sgalvao/runner/shared/config"
)

var client = &http.Client{Timeout: 15 * time.Second}

// File espelha o release.json do repositório: metadados do JRE por plataforma,
// a URL do assinador.jar e a URL/versão do simulador.jar.
type File struct {
	JRE struct {
		Version    string `json:"version"`
		WindowsX64 string `json:"windows_x64"`
		LinuxX64   string `json:"linux_x64"`
		MacX64     string `json:"mac_x64"`
	} `json:"jre"`
	Jar struct {
		URL string `json:"url"`
	} `json:"jar"`
	// Simulador descreve o simulador.jar externo: a Version controla a
	// invalidação do cache local (US-03.4); só rebaixa se a versão remota
	// diferir da gravada em ~/.hubsaude/simulador.version. SHA256 (hex),
	// quando presente, é verificado após o download.
	Simulador struct {
		URL     string `json:"url"`
		Version string `json:"version"`
		SHA256  string `json:"sha256"`
	} `json:"simulador"`
}

// Fetch baixa e desserializa o release.json a partir de config.ReleaseURL.
func Fetch() (*File, error) {
	resp, err := client.Get(config.ReleaseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d ao buscar release.json", resp.StatusCode)
	}
	var f File
	if err := json.NewDecoder(resp.Body).Decode(&f); err != nil {
		return nil, fmt.Errorf("release.json inválido: %w", err)
	}
	return &f, nil
}
