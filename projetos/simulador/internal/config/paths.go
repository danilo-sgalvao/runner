// Package config define os caminhos específicos do CLI simulador sob ~/.hubsaude,
// derivados da raiz de dados compartilhada (shared/config). Diferente do assinador,
// o registro de PID é gravado pelo próprio CLI Go (o simulador.jar é externo e não
// escreve em ~/.hubsaude), e o arquivo de versão controla o cache do download (US-03.4).
package config

import (
	"path/filepath"

	shared "github.com/danilo-sgalvao/runner/shared/config"
)

// JarPath retorna o caminho do simulador.jar gerenciado (~/.hubsaude/simulador.jar).
func JarPath() string {
	dir, _ := shared.HubSaudeDir()
	return filepath.Join(dir, "simulador.jar")
}

// PidPath retorna o caminho do registro de processo do simulador (~/.hubsaude/simulador.pid).
func PidPath() string {
	dir, _ := shared.HubSaudeDir()
	return filepath.Join(dir, "simulador.pid")
}

// VersionPath retorna o caminho do marcador de versão do jar baixado
// (~/.hubsaude/simulador.version), usado para invalidar o cache em US-03.4.
func VersionPath() string {
	dir, _ := shared.HubSaudeDir()
	return filepath.Join(dir, "simulador.version")
}
