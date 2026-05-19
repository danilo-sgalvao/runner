package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const releaseJSONURL = "https://raw.githubusercontent.com/danilo-sgalvao/runner/main/release.json"

type releaseConfig struct {
	Jar struct {
		URL string `json:"url"`
	} `json:"jar"`
}

// jarLocalPath retorna o caminho onde o jar gerenciado é armazenado (~/.hubsaude/assinador.jar).
func jarLocalPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".hubsaude", "assinador.jar")
}

// encontrarJar localiza o assinador.jar.
// Ordem de busca:
//  1. Mesma pasta do executável (modo distribuído)
//  2. ~/.hubsaude/assinador.jar (baixado automaticamente)
//  3. ../assinador-java/target/ (modo desenvolvimento local)
//  4. Download automático via release.json do repositório
func encontrarJar() (string, error) {
	if exe, err := os.Executable(); err == nil {
		jarAoLado := filepath.Join(filepath.Dir(exe), "assinador.jar")
		if _, err := os.Stat(jarAoLado); err == nil {
			return jarAoLado, nil
		}
	}

	if local := jarLocalPath(); fileExists(local) {
		return local, nil
	}

	dev := filepath.Join("..", "assinador-java", "target", "assinador.jar")
	if fileExists(dev) {
		return dev, nil
	}

	fmt.Println("assinador.jar não encontrado localmente. Baixando...")
	if err := downloadJar(); err != nil {
		return "", fmt.Errorf(
			"assinador.jar não encontrado e falha no download automático: %w\n"+
				"Em produção: coloque o assinador.jar na mesma pasta do executável.\n"+
				"Em desenvolvimento: execute 'mvn package' dentro de projetos/assinador-java/",
			err,
		)
	}

	local := jarLocalPath()
	if !fileExists(local) {
		return "", fmt.Errorf("download concluído mas assinador.jar não foi encontrado em %s", local)
	}
	return local, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func downloadJar() error {
	cfg, err := fetchReleaseConfig()
	if err != nil {
		return fmt.Errorf("não foi possível buscar release.json: %w", err)
	}
	if cfg.Jar.URL == "" {
		return fmt.Errorf("release.json não contém URL do assinador.jar")
	}

	resp, err := http.Get(cfg.Jar.URL) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("servidor retornou status %d ao baixar assinador.jar", resp.StatusCode)
	}

	dest := jarLocalPath()
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func fetchReleaseConfig() (*releaseConfig, error) {
	resp, err := http.Get(releaseJSONURL) //nolint:noctx
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("servidor retornou status %d ao buscar release.json", resp.StatusCode)
	}

	var cfg releaseConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("release.json inválido: %w", err)
	}
	return &cfg, nil
}
