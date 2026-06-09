package cmd

import (
	"net"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/danilo-sgalvao/runner/simulador/internal/simserver"
)

func isRegistered(name string) bool {
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == name {
			return true
		}
	}
	return false
}

func TestComandosRegistrados(t *testing.T) {
	for _, name := range []string{"version", "start", "stop", "status"} {
		if !isRegistered(name) {
			t.Errorf("comando %q não está registrado no rootCmd", name)
		}
	}
}

func TestRootUseCorreto(t *testing.T) {
	if rootCmd.Name() != "simulador" {
		t.Errorf("nome do comando raiz = %q, esperava \"simulador\"", rootCmd.Name())
	}
}

func TestStartFlags(t *testing.T) {
	port := startCmd.Flags().Lookup("port")
	if port == nil {
		t.Fatal("flag --port não encontrada em start")
	}
	if port.DefValue != strconv.Itoa(defaultPort) {
		t.Errorf("padrão de --port = %s, esperava %d", port.DefValue, defaultPort)
	}
	if startCmd.Flags().Lookup("source") == nil {
		t.Error("flag --source não encontrada em start")
	}
}

func TestStopStatusPortPadraoZero(t *testing.T) {
	for _, c := range []struct {
		name string
		flag string
	}{{"stop", "port"}, {"status", "port"}} {
		var f = stopCmd.Flags().Lookup(c.flag)
		if c.name == "status" {
			f = statusCmd.Flags().Lookup(c.flag)
		}
		if f == nil {
			t.Fatalf("flag --%s não encontrada em %s", c.flag, c.name)
		}
		if f.DefValue != "0" {
			t.Errorf("%s --%s padrão = %s, esperava 0", c.name, c.flag, f.DefValue)
		}
	}
}

func TestShortsNaoVazios(t *testing.T) {
	for _, c := range []*struct{ name, short string }{
		{"start", startCmd.Short},
		{"stop", stopCmd.Short},
		{"status", statusCmd.Short},
	} {
		if c.short == "" {
			t.Errorf("descrição curta de %q não deveria ser vazia", c.name)
		}
	}
}

// isolatePid aponta o registro de PID para um arquivo temporário inexistente.
func isolatePid(t *testing.T) {
	t.Helper()
	simserver.PidFilePath = filepath.Join(t.TempDir(), "simulador.pid")
	t.Cleanup(func() { simserver.PidFilePath = "" })
}

func TestStart_AbortaComPortaOcupada(t *testing.T) {
	isolatePid(t)

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("não foi possível ocupar uma porta: %v", err)
	}
	defer ln.Close()
	busy := ln.Addr().(*net.TCPAddr).Port

	startPort = busy
	startSource = ""
	t.Cleanup(func() { startPort = defaultPort })

	if err := startCmd.RunE(startCmd, nil); err == nil {
		t.Fatal("esperava erro ao iniciar com porta ocupada")
	}
}

func TestStatus_NaoEmExecucao(t *testing.T) {
	isolatePid(t)

	// Porta livre, nada respondendo → status reporta parado, sem erro.
	ln, _ := net.Listen("tcp", ":0")
	free := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	statusPort = free
	t.Cleanup(func() { statusPort = 0 })

	if err := statusCmd.RunE(statusCmd, nil); err != nil {
		t.Fatalf("status não deveria falhar quando o simulador está parado: %v", err)
	}
}

func TestStop_SemRegistro_RetornaErro(t *testing.T) {
	isolatePid(t)

	stopPort = 0
	if err := stopCmd.RunE(stopCmd, nil); err == nil {
		t.Fatal("esperava erro ao parar sem processo registrado")
	}
}
