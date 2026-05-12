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
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

const releaseURL = "https://raw.githubusercontent.com/danilo-sgalvao/runner/main/release.json"

type releaseFile struct {
	JRE struct {
		Version    string `json:"version"`
		WindowsX64 string `json:"windows_x64"`
		LinuxX64   string `json:"linux_x64"`
		MacX64     string `json:"mac_x64"`
	} `json:"jre"`
}

var javaVersionRe = regexp.MustCompile(`version "(\d+)`)

// JavaPath returns the absolute path to a java executable, provisioning the JRE if needed.
func JavaPath() (string, error) {
	if path, ok := localJREPath(); ok {
		return path, nil
	}

	if path, ok := systemJava(true); ok {
		return path, nil
	}

	rel, err := fetchRelease()
	if err != nil {
		// Offline fallback: use any java in PATH without version check
		if path, ok := systemJava(false); ok {
			fmt.Fprintln(os.Stderr, "Aviso: não foi possível verificar atualizações do JRE (sem rede). Usando java encontrado no sistema.")
			return path, nil
		}
		return "", fmt.Errorf("sem conexão e nenhum Java encontrado. Instale o Java 21 manualmente")
	}

	url, err := platformURL(rel)
	if err != nil {
		return "", err
	}

	return downloadAndInstall(url)
}

func hubsaudeJREDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".hubsaude", "jre"), nil
}

func javaExe() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

func localJREPath() (string, bool) {
	dir, err := hubsaudeJREDir()
	if err != nil {
		return "", false
	}
	p := filepath.Join(dir, "bin", javaExe())
	if _, err := os.Stat(p); err == nil {
		return p, true
	}
	return "", false
}

// systemJava looks for java in PATH. If requireV21 is true, only accepts Java 21+.
func systemJava(requireV21 bool) (string, bool) {
	path, err := exec.LookPath("java")
	if err != nil {
		return "", false
	}
	if !requireV21 {
		return path, true
	}
	out, err := exec.Command(path, "-version").CombinedOutput()
	if err != nil {
		return "", false
	}
	m := javaVersionRe.FindSubmatch(out)
	if m == nil {
		return "", false
	}
	major, err := strconv.Atoi(string(m[1]))
	if err != nil {
		return "", false
	}
	return path, major >= 21
}

func fetchRelease() (*releaseFile, error) {
	resp, err := http.Get(releaseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d ao buscar release.json", resp.StatusCode)
	}
	var rel releaseFile
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func platformURL(rel *releaseFile) (string, error) {
	switch runtime.GOOS {
	case "windows":
		return rel.JRE.WindowsX64, nil
	case "linux":
		return rel.JRE.LinuxX64, nil
	case "darwin":
		return rel.JRE.MacX64, nil
	default:
		return "", fmt.Errorf("plataforma não suportada: %s", runtime.GOOS)
	}
}

func downloadAndInstall(url string) (string, error) {
	fmt.Println("Java 21 não encontrado. Baixando JRE...")

	tmp, err := os.CreateTemp("", "hubsaude-jre-*")
	if err != nil {
		return "", fmt.Errorf("falha ao criar arquivo temporário: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := downloadWithProgress(url, tmp); err != nil {
		tmp.Close()
		return "", fmt.Errorf("falha no download: %w", err)
	}
	tmp.Close()

	destDir, err := hubsaudeJREDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("falha ao criar diretório %s: %w", destDir, err)
	}

	fmt.Println("Extraindo JRE...")
	if runtime.GOOS == "windows" {
		if err := extractZip(tmpPath, destDir); err != nil {
			return "", fmt.Errorf("falha ao extrair zip: %w", err)
		}
	} else {
		if err := extractTarGz(tmpPath, destDir); err != nil {
			return "", fmt.Errorf("falha ao extrair tar.gz: %w", err)
		}
	}

	javaPath := filepath.Join(destDir, "bin", javaExe())
	if _, err := os.Stat(javaPath); err != nil {
		return "", fmt.Errorf("java não encontrado após extração em %s", javaPath)
	}

	if runtime.GOOS != "windows" {
		_ = os.Chmod(javaPath, 0755)
	}

	fmt.Println("JRE instalado com sucesso em", destDir)
	return javaPath, nil
}

func downloadWithProgress(url string, dst *os.File) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	total := resp.ContentLength
	buf := make([]byte, 32*1024)
	var written int64

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return werr
			}
			written += int64(n)
			if total > 0 {
				fmt.Printf("\r  %.1f / %.1f MB (%.0f%%)",
					float64(written)/1e6, float64(total)/1e6,
					float64(written)/float64(total)*100)
			} else {
				fmt.Printf("\r  %.1f MB baixados", float64(written)/1e6)
			}
		}
		if readErr == io.EOF {
			fmt.Println()
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

// extractZip extracts a ZIP archive to destDir, stripping the top-level directory.
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	prefix := ""
	if len(r.File) > 0 {
		name := r.File[0].Name
		if i := strings.Index(name, "/"); i >= 0 {
			prefix = name[:i+1]
		}
	}

	for _, f := range r.File {
		name := strings.TrimPrefix(f.Name, prefix)
		if name == "" {
			continue
		}
		target := filepath.Join(destDir, filepath.FromSlash(name))

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, f.Mode()); err != nil {
				return err
			}
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

// extractTarGz extracts a .tar.gz archive to destDir, stripping the top-level directory.
func extractTarGz(tarPath, destDir string) error {
	f, err := os.Open(tarPath)
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
	prefix := ""

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if prefix == "" {
			if i := strings.Index(hdr.Name, "/"); i >= 0 {
				prefix = hdr.Name[:i+1]
			}
		}

		name := strings.TrimPrefix(hdr.Name, prefix)
		if name == "" {
			continue
		}
		target := filepath.Join(destDir, filepath.FromSlash(name))

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, hdr.FileInfo().Mode()); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, hdr.FileInfo().Mode())
			if err != nil {
				return err
			}
			_, err = io.Copy(out, tr)
			out.Close()
			if err != nil {
				return err
			}
		case tar.TypeSymlink:
			_ = os.Symlink(hdr.Linkname, target)
		}
	}
	return nil
}
