package util

import (
	"github.com/akawashiro/go-common/src/utils"
	"syscall"
)

// StartEmergencyShell exec's /bin/sh.
func StartEmergencyShell() {
	Say("Starting emergency shell...")

	shell := "/bin/sh"
	syscall.Exec(shell, utils.StringSlice(shell), nil)
}
