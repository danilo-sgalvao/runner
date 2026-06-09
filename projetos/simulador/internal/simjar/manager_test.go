package simjar

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/danilo-sgalvao/runner/shared/release"
	"github.com/danilo-sgalvao/runner/simulador/internal/config"
)

// isolateHome aponta ~/.hubsaude para um diretório temporário e neutraliza o
// atalho de "jar ao lado do executável", para que cada teste parta de um estado
// limpo e previsível.
func isolateHome(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("USERPROFILE", tmp) // Windows
	t.Setenv("HOME", tmp)        // Unix
	if exe, err := os.Executable(); err == nil {
		os.Remove(filepath.Join(filepath.Dir(exe), "simulador.jar"))
	}
}

// stubRelease substitui fetchRelease durante o teste e o restaura ao fim.
func stubRelease(t *testing.T, f *release.File, err error) {
	t.Helper()
	orig := fetchRelease
	fetchRelease = func() (*release.File, error) { return f, err }
	t.Cleanup(func() { fetchRelease = orig })
}

func releaseWith(url, version string) *release.File {
	f := &release.File{}
	f.Simulador.URL = url
	f.Simulador.Version = version
	return f
}

// jarServer serve bytes fixos como se fosse o jar.
func jarServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func TestFind_DownloadFresh(t *testing.T) {
	isolateHome(t)
	srv := jarServer(t, "jar-fresco")
	stubRelease(t, releaseWith(srv.URL, "1.0.0"), nil)

	path, err := Find("")
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if path != config.JarPath() {
		t.Errorf("esperava %s, obteve %s", config.JarPath(), path)
	}
	if got := readFile(t, path); got != "jar-fresco" {
		t.Errorf("conteúdo do jar = %q, esperava %q", got, "jar-fresco")
	}
	if got := readFile(t, config.VersionPath()); got != "1.0.0" {
		t.Errorf("marcador de versão = %q, esperava %q", got, "1.0.0")
	}
}

func TestFind_CacheValido(t *testing.T) {
	isolateHome(t)
	writeFile(t, config.JarPath(), "jar-em-cache")
	writeFile(t, config.VersionPath(), "1.0.0")

	// Servidor que falha o teste se for acionado: cache válido não deve baixar.
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("não deveria baixar com cache válido")
	}))
	t.Cleanup(srv.Close)
	stubRelease(t, releaseWith(srv.URL, "1.0.0"), nil)

	path, err := Find("")
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if got := readFile(t, path); got != "jar-em-cache" {
		t.Errorf("jar foi rebaixado; conteúdo = %q", got)
	}
}

func TestFind_CacheDesatualizado(t *testing.T) {
	isolateHome(t)
	writeFile(t, config.JarPath(), "jar-antigo")
	writeFile(t, config.VersionPath(), "0.9.0")
	srv := jarServer(t, "jar-novo")
	stubRelease(t, releaseWith(srv.URL, "1.0.0"), nil)

	path, err := Find("")
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if got := readFile(t, path); got != "jar-novo" {
		t.Errorf("esperava jar atualizado, obteve %q", got)
	}
	if got := readFile(t, config.VersionPath()); got != "1.0.0" {
		t.Errorf("marcador de versão não atualizado: %q", got)
	}
}

func TestFind_SourceSobrepoeRelease(t *testing.T) {
	isolateHome(t)
	srv := jarServer(t, "jar-de-source")
	// fetchRelease não deve ser chamado quando --source é informado.
	stubRelease(t, nil, http.ErrServerClosed)
	orig := fetchRelease
	fetchRelease = func() (*release.File, error) {
		t.Error("fetchRelease não deveria ser chamado com --source")
		return orig()
	}
	t.Cleanup(func() { fetchRelease = orig })

	path, err := Find(srv.URL)
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if got := readFile(t, path); got != "jar-de-source" {
		t.Errorf("esperava jar de --source, obteve %q", got)
	}
}

func TestFind_OfflineSemCache(t *testing.T) {
	isolateHome(t)
	stubRelease(t, nil, http.ErrServerClosed)

	_, err := Find("")
	if err == nil {
		t.Fatal("esperava erro quando offline e sem cache")
	}
}

func TestFind_OfflineComCache(t *testing.T) {
	isolateHome(t)
	writeFile(t, config.JarPath(), "jar-em-cache")
	stubRelease(t, nil, http.ErrServerClosed)

	path, err := Find("")
	if err != nil {
		t.Fatalf("não esperava erro com cache presente: %v", err)
	}
	if got := readFile(t, path); got != "jar-em-cache" {
		t.Errorf("esperava usar o cache, obteve %q", got)
	}
}
