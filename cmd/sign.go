package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var signContent string
var signAlgorithm string

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Cria uma assinatura digital simulada",
	Run: func(cmd *cobra.Command, args []string) {
		if signContent == "" {
			fmt.Println("Erro: --content é obrigatório.")
			os.Exit(1)
		}

		jarPath := encontrarJar()

		javaCmd := exec.Command("java", "-jar", jarPath,
			"sign",
			"--content", signContent,
			"--algorithm", signAlgorithm,
		)

		javaCmd.Stdout = os.Stdout
		javaCmd.Stderr = os.Stderr

		err := javaCmd.Run()
		if err != nil {
			fmt.Println("Erro ao executar assinador.jar:", err)
			os.Exit(1)
		}
	},
}

func encontrarJar() string {
	// Procura o jar na mesma pasta do executável
	exe, err := os.Executable()
	if err == nil {
		jarAoLado := filepath.Join(filepath.Dir(exe), "assinador.jar")
		if _, err := os.Stat(jarAoLado); err == nil {
			return jarAoLado
		}
	}

	// Procura na pasta local (desenvolvimento)
	local := filepath.Join("assinador", "target", "assinador.jar")
	if _, err := os.Stat(local); err == nil {
		return local
	}

	fmt.Println("Erro: assinador.jar não encontrado.")
	if runtime.GOOS == "windows" {
		fmt.Println("Coloque o assinador.jar na mesma pasta do executável.")
	}
	os.Exit(1)
	return ""
}

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.Flags().StringVar(&signContent, "content", "", "Conteúdo a ser assinado (obrigatório)")
	signCmd.Flags().StringVar(&signAlgorithm, "algorithm", "SHA256withRSA", "Algoritmo de assinatura")
}