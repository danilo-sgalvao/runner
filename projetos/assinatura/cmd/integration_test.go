//go:build integration

package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/danilo-sgalvao/runner/internal/jre"
)

// jarDeIntegracao resolve o caminho do assinador.jar para testes.
// go test ./cmd roda com cwd = projetos/assinatura/cmd/, então subimos dois níveis.
func jarDeIntegracao(t *testing.T) string {
	t.Helper()
	jarPath := filepath.Join("..", "..", "assinador-java", "target", "assinador.jar")
	if _, err := os.Stat(jarPath); err != nil {
		t.Skipf("assinador.jar não encontrado em %s — execute 'mvn package' em projetos/assinador-java/ antes de rodar testes de integração", filepath.Clean(jarPath))
	}
	return jarPath
}

func javaDeIntegracao(t *testing.T) string {
	t.Helper()
	javaPath, err := jre.JavaPath()
	if err != nil {
		t.Skipf("Java não disponível: %v", err)
	}
	return javaPath
}

func TestIntegracao_Sign_RetornaJsonComSucesso(t *testing.T) {
	jar := jarDeIntegracao(t)
	java := javaDeIntegracao(t)

	out, err := exec.Command(java, "-jar", jar, "sign", "--content", "documento de teste").Output()
	if err != nil {
		t.Fatalf("execução do JAR falhou: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(out))), &resp); err != nil {
		t.Fatalf("saída não é JSON válido: %v\nsaída: %s", err, out)
	}

	if valid, _ := resp["valid"].(bool); !valid {
		t.Errorf("esperava valid=true, obteve: %v", resp["valid"])
	}
	if sig, _ := resp["signature"].(string); sig == "" {
		t.Errorf("campo signature ausente ou vazio")
	}
}

func TestIntegracao_Validate_AssinaturaCorreta_RetornaValida(t *testing.T) {
	jar := jarDeIntegracao(t)
	java := javaDeIntegracao(t)

	out, err := exec.Command(java, "-jar", jar,
		"validate",
		"--content", "documento",
		"--signature", "MOCKED_SIGNATURE_BASE64_==",
	).Output()
	if err != nil {
		t.Fatalf("execução do JAR falhou: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(out))), &resp); err != nil {
		t.Fatalf("saída não é JSON válido: %v\nsaída: %s", err, out)
	}

	if valid, _ := resp["valid"].(bool); !valid {
		t.Errorf("esperava valid=true para assinatura correta, obteve: %v", resp["valid"])
	}
}

func TestIntegracao_Validate_AssinaturaErrada_RetornaInvalidaComExitCode1(t *testing.T) {
	jar := jarDeIntegracao(t)
	java := javaDeIntegracao(t)

	var stderr bytes.Buffer
	cmd := exec.Command(java, "-jar", jar,
		"validate",
		"--content", "documento",
		"--signature", "ASSINATURA-ERRADA",
	)
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err == nil {
		t.Fatal("esperava exit code != 0 para assinatura inválida")
	}

	var resp map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stderr.String())), &resp); jsonErr != nil {
		t.Fatalf("stderr não é JSON válido: %v\nstderr: %s", jsonErr, stderr.String())
	}

	if valid, _ := resp["valid"].(bool); valid {
		t.Errorf("esperava valid=false para assinatura errada, obteve true")
	}
}

func TestIntegracao_Sign_ContentVazio_RetornaErroComExitCode1(t *testing.T) {
	jar := jarDeIntegracao(t)
	java := javaDeIntegracao(t)

	err := exec.Command(java, "-jar", jar, "sign", "--content", "").Run()
	if err == nil {
		t.Fatal("esperava exit code != 0 para content vazio")
	}
}

func TestIntegracao_FluxoCompleto_SignEntaoValidate(t *testing.T) {
	jar := jarDeIntegracao(t)
	java := javaDeIntegracao(t)

	// sign
	out, err := exec.Command(java, "-jar", jar, "sign", "--content", "doc importante").Output()
	if err != nil {
		t.Fatalf("sign falhou: %v", err)
	}
	var signResp map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(out))), &signResp); err != nil {
		t.Fatalf("saída do sign não é JSON válido: %v", err)
	}
	sig, _ := signResp["signature"].(string)
	if sig == "" {
		t.Fatal("sign não retornou signature")
	}

	// validate com a assinatura obtida
	out, err = exec.Command(java, "-jar", jar,
		"validate", "--content", "doc importante", "--signature", sig,
	).Output()
	if err != nil {
		t.Fatalf("validate falhou: %v", err)
	}
	var valResp map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(out))), &valResp); err != nil {
		t.Fatalf("saída do validate não é JSON válido: %v", err)
	}
	if valid, _ := valResp["valid"].(bool); !valid {
		t.Errorf("validate deveria aprovar assinatura retornada pelo sign")
	}
}
