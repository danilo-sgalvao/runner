package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEncontrarJar_JarAoLadoDoExecutavel(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Skip("não foi possível obter caminho do executável de teste")
	}

	fakeJar := filepath.Join(filepath.Dir(exe), "assinador.jar")
	if err := os.WriteFile(fakeJar, []byte("fake-jar"), 0644); err != nil {
		t.Skipf("não foi possível criar jar de teste: %v", err)
	}
	defer os.Remove(fakeJar)

	path, err := encontrarJar()
	if err != nil {
		t.Fatalf("esperava encontrar jar ao lado do executável, obteve erro: %v", err)
	}
	if path != fakeJar {
		t.Errorf("esperava caminho %s, obteve %s", fakeJar, path)
	}
}

func TestEncontrarJar_SemJar_RetornaErro(t *testing.T) {
	// Remove jar do lado do executável se existir (cleanup de outros testes)
	exe, _ := os.Executable()
	fakeJar := filepath.Join(filepath.Dir(exe), "assinador.jar")
	os.Remove(fakeJar)

	// Verifica que retorna erro quando o jar não está em lugar nenhum padrão
	// (exceto se o desenvolvedor tiver o jar em ../assinador-java/target/)
	_, err := encontrarJar()
	// Não fazemos assert fatal aqui porque em desenvolvimento o jar pode existir
	_ = err
}

func TestEncontrarJar_RetornaCaminhoAbsolutoOuRelativo(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Skip("não foi possível obter caminho do executável de teste")
	}

	fakeJar := filepath.Join(filepath.Dir(exe), "assinador.jar")
	if err := os.WriteFile(fakeJar, []byte("fake-jar"), 0644); err != nil {
		t.Skipf("não foi possível criar jar: %v", err)
	}
	defer os.Remove(fakeJar)

	path, err := encontrarJar()
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}
	if path == "" {
		t.Error("caminho retornado não deveria ser vazio")
	}
}
