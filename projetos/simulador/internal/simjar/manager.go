// Package simjar localiza e obtém dinamicamente o simulador.jar (US-03.4).
//
// Diferente do assinador.jar (construído neste repositório), o simulador.jar é um
// artefato externo: ele só é baixado, nunca compilado localmente. Por isso o
// download é versionado — grava-se ~/.hubsaude/simulador.version ao baixar e
// compara-se com a versão anunciada no release.json antes de baixar de novo, para
// não rebaixar o que já está atualizado.
package simjar

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/danilo-sgalvao/runner/shared/release"
	"github.com/danilo-sgalvao/runner/simulador/internal/config"
)

// fetchRelease é um ponto de injeção para testes; em produção lê o release.json real.
var fetchRelease = release.Fetch

// Find localiza o simulador.jar, baixando-o quando necessário.
//
// sourceURL, quando não vazio, sobrepõe a URL do release.json (flag --source) e
// força o download a partir dela, ignorando o cache por versão.
//
// Ordem de resolução:
//  1. Atalho: simulador.jar ao lado do executável (modo distribuído).
//  2. --source informado → baixa dessa URL (sempre).
//  3. release.json: se ~/.hubsaude/simulador.jar existe e simulador.version == versão
//     remota, usa o cache; senão baixa e regrava jar + versão.
//  4. Offline (falha ao ler release.json) com cache presente → usa o cache.
//  5. Offline sem cache → erro claro.
func Find(sourceURL string) (string, error) {
	if exe, err := os.Executable(); err == nil {
		jarAoLado := filepath.Join(filepath.Dir(exe), "simulador.jar")
		if fileExists(jarAoLado) {
			return jarAoLado, nil
		}
	}

	if sourceURL != "" {
		fmt.Println("Baixando simulador.jar de --source...")
		if err := download(sourceURL, ""); err != nil {
			return "", fmt.Errorf("falha ao baixar simulador.jar de %s: %w", sourceURL, err)
		}
		return config.JarPath(), nil
	}

	local := config.JarPath()
	cfg, err := fetchRelease()
	if err != nil {
		// Sem rede: o cache local, se existir, ainda é utilizável.
		if fileExists(local) {
			return local, nil
		}
		return "", fmt.Errorf(
			"simulador.jar não encontrado localmente e não foi possível consultar o release.json: %w\n"+
				"Conecte-se à internet para o primeiro download ou use --source <url>.",
			err,
		)
	}
	if cfg.Simulador.URL == "" {
		return "", fmt.Errorf("release.json não contém a URL do simulador.jar")
	}

	if fileExists(local) && localVersion() == cfg.Simulador.Version {
		return local, nil
	}

	fmt.Printf("Baixando simulador.jar (versão %s)...\n", cfg.Simulador.Version)
	if err := download(cfg.Simulador.URL, cfg.Simulador.Version); err != nil {
		// Falha no download mas com cache presente: degrada para o cache.
		if fileExists(local) {
			fmt.Println("Aviso: falha ao atualizar; usando a cópia local existente.")
			return local, nil
		}
		return "", fmt.Errorf("falha ao baixar simulador.jar: %w", err)
	}
	return local, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// localVersion lê o marcador de versão do jar em cache; "" se ausente.
func localVersion() string {
	data, err := os.ReadFile(config.VersionPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// download baixa o jar de url para ~/.hubsaude/simulador.jar. Se version não for
// vazio, grava também o marcador de versão para o cache de US-03.4.
func download(url, version string) error {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("servidor retornou status %d", resp.StatusCode)
	}

	dest := config.JarPath()
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	// Grava em arquivo temporário e renomeia, para não deixar um jar parcial
	// no destino caso o download seja interrompido.
	tmp := dest + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return err
	}

	if version != "" {
		if err := os.WriteFile(config.VersionPath(), []byte(version), 0644); err != nil {
			return fmt.Errorf("jar baixado mas não foi possível gravar o marcador de versão: %w", err)
		}
	}
	return nil
}
