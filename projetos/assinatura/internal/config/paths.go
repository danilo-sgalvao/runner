// Package config centraliza os caminhos da aplicação sob ~/.hubsaude e a URL do
// metadado de release, servindo de fonte única para os pacotes jar, jre e server.
package config

import (
	"os"
	"path/filepath"
)

// ReleaseURL é a fonte única do metadado de release (JRE + jars) do projeto.
const ReleaseURL = "https://raw.githubusercontent.com/danilo-sgalvao/runner/main/release.json"

// dirName é o único lugar onde o nome do diretório de dados é definido.
const dirName = ".hubsaude"

// HubSaudeDir retorna a raiz de dados da aplicação (~/.hubsaude).
func HubSaudeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, dirName), nil
}

// JarPath retorna o caminho do assinador.jar gerenciado (~/.hubsaude/assinador.jar).
func JarPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, dirName, "assinador.jar")
}

// PidPath retorna o caminho do registro de processo do servidor (~/.hubsaude/assinador.pid).
func PidPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, dirName, "assinador.pid")
}

// JREDir retorna o diretório do JRE gerenciado (~/.hubsaude/jre).
func JREDir() (string, error) {
	dir, err := HubSaudeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "jre"), nil
}
