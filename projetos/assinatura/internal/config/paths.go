// Package config define os caminhos específicos do CLI assinatura sob ~/.hubsaude,
// derivados da raiz de dados compartilhada (shared/config).
package config

import (
	"path/filepath"

	shared "github.com/danilo-sgalvao/runner/shared/config"
)

// JarPath retorna o caminho do assinador.jar gerenciado (~/.hubsaude/assinador.jar).
func JarPath() string {
	dir, _ := shared.HubSaudeDir()
	return filepath.Join(dir, "assinador.jar")
}

// PidPath retorna o caminho do registro de processo do servidor (~/.hubsaude/assinador.pid).
func PidPath() string {
	dir, _ := shared.HubSaudeDir()
	return filepath.Join(dir, "assinador.pid")
}
