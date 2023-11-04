//go:build windows && (x86 || 386)
// +build windows
// +build x86 386

package wintun

import (
	"embed"
)

//go:embed bin/x86/wintun.dll
var wintunFs embed.FS

func InstallWintunDriver() error {
	bytes, err := wintunFs.ReadFile("bin/x86/wintun.dll")
	if err != nil {
		err = errors.New("wintunFs.ReadFile("bin/x86/wintun.dll"): " + err.Error())
		return err
	}
	return copyDriver(bytes)
}
