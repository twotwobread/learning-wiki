package hostpath

import (
	"os/exec"
	"runtime"
)

// ForCurrentHost enables Windows-drive input only when the WSL path utility is available.
// Linux paths remain valid on ordinary Linux installations without WSL.
func ForCurrentHost() Resolver {
	if runtime.GOOS == "linux" {
		if _, err := exec.LookPath("wslpath"); err == nil {
			return WSL(nil)
		}
	}
	return Native()
}
