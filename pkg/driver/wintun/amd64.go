//go:build windows && amd64
// +build windows,amd64

package wintun

import (
	"embed"
)

//go:embed bin/amd64/wintun.dll
var wintunFs embed.FS

func InstallWintunDriver() error {
	bytes, err := wintunFs.ReadFile("bin/amd64/wintun.dll")
	if err != nil {
		err = errors.New("wintunFs.ReadFile("bin/amd64/wintun.dll"): " + err.Error())
		return err
	}
	return copyDriver(bytes)
}
