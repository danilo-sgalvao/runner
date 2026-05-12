// Package jre detecta, baixa e configura automaticamente um JRE compatível,
// implementando o fluxo definido em docs/plano-download-java.md (US-04.1).
package jre

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const releaseJSONURL = "https://raw.githubusercontent.com/danilo-sgalvao/runner/main/release.json"

type releaseConfig struct {
	JRE struct {
		Version    string `json:"version"`
		WindowsX64 string `json:"windows_x64"`
		LinuxX64   string `json:"linux_x64"`
		MacX64     string `json:"mac_x64"`
	} `json:"jre"`
}

// JavaPath retorna o caminho absoluto do executável java pronto para uso.
//
// Prioridade de busca:
//  1. JRE gerenciado localmente em ~/.hubsaude/jre  (sem overhead de rede)
//  2. java disponível no PATH do sistema
//  3. Download automático via release.json do repositório
func JavaPath() (string, error) {
	if path, ok := detectLocal(); ok {
		return path, nil
	}

	if path, ok := detectSystem(); ok {
		return path, nil
	}

	cfg, err := fetchReleaseJSON()
	if err != nil {
		// Sem acesso à internet: tenta qualquer java do PATH como fallback
		if path, err2 := exec.LookPath("java"); err2 == nil {
			fmt.Fprintf(os.Stderr,
				"Aviso: não foi possível verificar versão do Java (%v).\nUsando java do PATH como fallback.\n", err)
			return path, nil
		}
		return "", fmt.Errorf(
			"Java não encontrado e sem acesso à internet para baixar.\n"+
				"Instale o Java 21+ manualmente ou conecte-se à internet: %w", err)
	}

	url := jreURL(cfg)
	if url == "" {
		return "", fmt.Errorf(
			"plataforma não suportada para download automático: %s/%s.\n"+
				"Instale o Java 21+ manualmente.", runtime.GOOS, runtime.GOARCH)
	}

	if err := downloadAndExtract(url); err != nil {
		return "", fmt.Errorf("falha ao baixar JRE: %w", err)
	}

	path, ok := detectLocal()
	if !ok {
		return "", fmt.Errorf("JRE baixado mas executável java não foi encontrado em %s", LocalJREDir())
	}
	return path, nil
}

// LocalJREDir retorna o diretório onde o JRE gerenciado é armazenado.
func LocalJREDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".hubsaude", "jre")
}

func detectLocal() (string, bool) {
	var exe string
	if runtime.GOOS == "windows" {
		exe = filepath.Join(LocalJREDir(), "bin", "java.exe")
	} else {
		exe = filepath.Join(LocalJREDir(), "bin", "java")
	}
	if _, err := os.Stat(exe); err == nil {
		return exe, true
	}
	return "", false
}

func detectSystem() (string, bool) {
	path, err := exec.LookPath("java")
	if err != nil {
		return "", false
	}
	return path, true
}

func fetchReleaseJSON() (*releaseConfig, error) {
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

func jreURL(cfg *releaseConfig) string {
	switch {
	case runtime.GOOS == "windows" && runtime.GOARCH == "amd64":
		return cfg.JRE.WindowsX64
	case runtime.GOOS == "linux" && runtime.GOARCH == "amd64":
		return cfg.JRE.LinuxX64
	case runtime.GOOS == "darwin":
		return cfg.JRE.MacX64
	default:
		return ""
	}
}

func downloadAndExtract(url string) error {
	fmt.Println("Baixando Java (JRE)... isso pode levar alguns instantes.")

	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tmp, err := os.CreateTemp("", "jre-download-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	dest := LocalJREDir()
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	fmt.Println("Extraindo JRE...")
	if runtime.GOOS == "windows" {
		return extractZIP(tmp.Name(), dest)
	}
	return extractTarGZ(tmp.Name(), dest)
}

// extractZIP extrai um arquivo .zip, ignorando o diretório raiz do pacote.
func extractZIP(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Remove o diretório raiz do JRE (ex.: jdk-21.0.x-jre/)
		parts := strings.SplitN(f.Name, "/", 2)
		if len(parts) < 2 || parts[1] == "" {
			continue
		}
		target := filepath.Join(dest, filepath.FromSlash(parts[1]))

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, f.Mode())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// extractTarGZ extrai um arquivo .tar.gz, ignorando o diretório raiz do pacote.
func extractTarGZ(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Remove o diretório raiz do JRE (ex.: jdk-21.0.x-jre/)
		parts := strings.SplitN(hdr.Name, "/", 2)
		if len(parts) < 2 || parts[1] == "" {
			continue
		}
		target := filepath.Join(dest, parts[1])

		switch hdr.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, os.FileMode(hdr.Mode))
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			_, err = io.Copy(out, tr)
			out.Close()
			if err != nil {
				return err
			}
		case tar.TypeSymlink:
			os.Symlink(hdr.Linkname, target)
		}
	}
	return nil
}
