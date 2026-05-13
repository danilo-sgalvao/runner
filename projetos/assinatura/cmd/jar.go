package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

// encontrarJar localiza o assinador.jar.
// Ordem de busca:
//  1. Mesma pasta do executável (modo distribuído)
//  2. ../assinador-java/target/ (modo desenvolvimento)
func encontrarJar() (string, error) {
	exe, err := os.Executable()
	if err == nil {
		jarAoLado := filepath.Join(filepath.Dir(exe), "assinador.jar")
		if _, err := os.Stat(jarAoLado); err == nil {
			return jarAoLado, nil
		}
	}

	local := filepath.Join("..", "assinador-java", "target", "assinador.jar")
	if _, err := os.Stat(local); err == nil {
		return local, nil
	}

	return "", fmt.Errorf(
		"assinador.jar não encontrado.\n" +
			"Em produção: coloque o assinador.jar na mesma pasta do executável.\n" +
			"Em desenvolvimento: execute 'mvn package' dentro de projetos/assinador-java/",
	)
}
