package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/danilo-sgalvao/runner/assinatura/internal/jar"
	"github.com/danilo-sgalvao/runner/assinatura/internal/server"
	"github.com/danilo-sgalvao/runner/shared/jre"
)

// runViaServer tenta atender a operação pelo servidor HTTP em execução.
//
// Retorna handled=true quando o servidor foi usado — nesse caso o erro retornado
// (se houver) é o da chamada HTTP e o chamador deve repassá-lo sem cair para o
// modo local. Retorna handled=false quando --local foi pedido ou não há servidor
// respondendo, sinalizando ao chamador que use runViaJar.
func runViaServer(local bool, call func(port int) (*server.SignatureResponse, error)) (handled bool, err error) {
	if local {
		return false, nil
	}
	info, err := server.ReadProcessInfo()
	if err != nil || !server.IsResponding(info.Port) {
		return false, nil
	}
	resp, err := call(info.Port)
	if err != nil {
		return true, fmt.Errorf("erro ao chamar servidor: %w", err)
	}
	fmt.Printf("Assinatura: %s\nVálido: %v\nMensagem: %s\n",
		resp.Signature, resp.Valid, resp.Message)
	return true, nil
}

// runViaJar executa o assinador.jar localmente via java -jar, herdando stdout/stderr.
func runViaJar(args ...string) error {
	jarPath, err := jar.Find()
	if err != nil {
		return err
	}

	javaPath, err := jre.JavaPath()
	if err != nil {
		return fmt.Errorf("Java não disponível: %w", err)
	}

	javaCmd := exec.Command(javaPath, append([]string{"-jar", jarPath}, args...)...)
	javaCmd.Stdout = os.Stdout
	javaCmd.Stderr = os.Stderr
	return javaCmd.Run()
}
