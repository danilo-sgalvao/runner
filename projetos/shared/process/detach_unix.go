//go:build !windows

package process

import (
	"os/exec"
	"syscall"
)

// Detach configura o comando para rodar em nova sessão, sobrevivendo ao encerramento do CLI.
func Detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
