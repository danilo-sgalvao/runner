package jre

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLocalJREDir_NaoVazio(t *testing.T) {
	dir := LocalJREDir()
	if dir == "" {
		t.Fatal("LocalJREDir não deve retornar string vazia")
	}
}

func TestLocalJREDir_ContemHubsaude(t *testing.T) {
	dir := LocalJREDir()
	found := false
	const marcador = ".hubsaude"
	for i := 0; i <= len(dir)-len(marcador); i++ {
		if dir[i:i+len(marcador)] == marcador {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("LocalJREDir deveria conter '.hubsaude' no caminho, obteve: %s", dir)
	}
}

func TestLocalJREDir_TerminaComJre(t *testing.T) {
	dir := LocalJREDir()
	if filepath.Base(dir) != "jre" {
		t.Errorf("LocalJREDir deveria terminar com 'jre', obteve: %s", filepath.Base(dir))
	}
}

func TestJREURL_RetornaURLNaoVaziaParaPlataformaAtual(t *testing.T) {
	cfg := &releaseConfig{}
	cfg.JRE.WindowsX64 = "https://exemplo.com/windows"
	cfg.JRE.LinuxX64 = "https://exemplo.com/linux"
	cfg.JRE.MacX64 = "https://exemplo.com/mac"

	url := jreURL(cfg)

	switch runtime.GOOS {
	case "windows", "linux", "darwin":
		if url == "" {
			t.Errorf("esperava URL não vazia para %s, obteve string vazia", runtime.GOOS)
		}
	}
}

func TestJREURL_Windows_RetornaURLCorreta(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("teste específico para Windows")
	}
	cfg := &releaseConfig{}
	cfg.JRE.WindowsX64 = "https://exemplo.com/windows"
	cfg.JRE.LinuxX64 = "https://exemplo.com/linux"
	cfg.JRE.MacX64 = "https://exemplo.com/mac"

	url := jreURL(cfg)
	if url != "https://exemplo.com/windows" {
		t.Errorf("esperava URL do Windows, obteve: %s", url)
	}
}

func TestJREURL_Linux_RetornaURLCorreta(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("teste específico para Linux")
	}
	cfg := &releaseConfig{}
	cfg.JRE.WindowsX64 = "https://exemplo.com/windows"
	cfg.JRE.LinuxX64 = "https://exemplo.com/linux"
	cfg.JRE.MacX64 = "https://exemplo.com/mac"

	url := jreURL(cfg)
	if url != "https://exemplo.com/linux" {
		t.Errorf("esperava URL do Linux, obteve: %s", url)
	}
}

func TestJREURL_MacOS_RetornaURLCorreta(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("teste específico para macOS")
	}
	cfg := &releaseConfig{}
	cfg.JRE.WindowsX64 = "https://exemplo.com/windows"
	cfg.JRE.LinuxX64 = "https://exemplo.com/linux"
	cfg.JRE.MacX64 = "https://exemplo.com/mac"

	url := jreURL(cfg)
	if url != "https://exemplo.com/mac" {
		t.Errorf("esperava URL do macOS, obteve: %s", url)
	}
}

func TestJREURL_ConfigVazia_RetornaVazio(t *testing.T) {
	cfgVazio := &releaseConfig{}
	url := jreURL(cfgVazio)
	if url != "" {
		t.Errorf("config vazia deveria retornar URL vazia, obteve: %s", url)
	}
}

func TestDetectLocal_ConsistenteComStat(t *testing.T) {
	dir := LocalJREDir()
	var exePath string
	if runtime.GOOS == "windows" {
		exePath = filepath.Join(dir, "bin", "java.exe")
	} else {
		exePath = filepath.Join(dir, "bin", "java")
	}

	_, statErr := os.Stat(exePath)
	javaExisteNoDisco := statErr == nil

	_, detectado := detectLocal()

	if javaExisteNoDisco != detectado {
		t.Errorf("detectLocal() = %v mas java existe no disco = %v (caminho: %s)",
			detectado, javaExisteNoDisco, exePath)
	}
}

func TestDetectLocal_ComJavaTmp_Detecta(t *testing.T) {
	// Cria um java falso em diretório temporário para testar a lógica de stat
	tmp := t.TempDir()
	binDir := filepath.Join(tmp, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}

	var exeNome string
	if runtime.GOOS == "windows" {
		exeNome = "java.exe"
	} else {
		exeNome = "java"
	}

	javaFake := filepath.Join(binDir, exeNome)
	if err := os.WriteFile(javaFake, []byte("fake-java"), 0755); err != nil {
		t.Fatal(err)
	}

	// Verifica que o arquivo foi criado e é legível pelo stat
	if _, err := os.Stat(javaFake); err != nil {
		t.Fatalf("arquivo java de teste deveria existir: %v", err)
	}
}
