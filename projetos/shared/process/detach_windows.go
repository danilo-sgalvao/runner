//go:build windows

package process

import (
	"os/exec"
	"syscall"
)

// Detach configura o comando para rodar como novo grupo de processos,
// sobrevivendo ao encerramento do CLI no Windows.
func Detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
