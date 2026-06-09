// Package config centraliza os caminhos genéricos da aplicação sob ~/.hubsaude e
// a URL do metadado de release, servindo de fonte única compartilhada pelos CLIs
// assinatura e simulador. Caminhos específicos de cada CLI (ex.: assinador.jar,
// simulador.pid) vivem no pacote config de cada módulo.
package config

import (
	"os"
	"path/filepath"
)

// ReleaseURL é a fonte única do metadado de release (JRE + jars) do projeto.
const ReleaseURL = "https://raw.githubusercontent.com/danilo-sgalvao/runner/main/release.json"

// DirName é o nome do diretório de dados da aplicação (~/.hubsaude). Exportado
// para que os pacotes config de cada CLI derivem seus próprios caminhos.
const DirName = ".hubsaude"

// HubSaudeDir retorna a raiz de dados da aplicação (~/.hubsaude).
func HubSaudeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, DirName), nil
}

// JREDir retorna o diretório do JRE gerenciado (~/.hubsaude/jre).
func JREDir() (string, error) {
	dir, err := HubSaudeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "jre"), nil
}
