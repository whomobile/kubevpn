package upgrade

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	goversion "github.com/hashicorp/go-version"

	"github.com/wencaiwulue/kubevpn/pkg/config"
	"github.com/wencaiwulue/kubevpn/pkg/util"
)

// Main
// 1) get current binary version
// 2) get the latest version
// 3) compare two version decide needs to download or not
// 4) download newer version zip
// 5) unzip to temp file
// 6) check permission of putting new kubevpn back
// 7) chmod +x, move old to /temp, move new to CURRENT_FOLDER
func Main(ctx context.Context, client *http.Client, latestVersion string, latestCommit string, url string) error {
	fmt.Printf("The latest version is: %s, commit: %s\n", latestVersion, latestCommit)

	commit := config.GitCommit
	version := config.Version
	var err error
	var cVersion, dVersion *goversion.Version
	cVersion, err = goversion.NewVersion(version)
	if err != nil {
		err = errors.New("goversion.NewVersion(version): " + err.Error())
		return err
	}
	dVersion, err = goversion.NewVersion(latestVersion)
	if err != nil {
		err = errors.New("goversion.NewVersion(latestVersion): " + err.Error())
		return err
	}
	if cVersion.GreaterThan(dVersion) || (cVersion.Equal(dVersion) && commit == latestCommit) {
		fmt.Println("Already up to date, don't needs to upgrade")
		return nil
	}

	var executable string
	executable, err = os.Executable()
	if err != nil {
		err = errors.New("os.Executable(): " + err.Error())
		return err
	}
	var tem *os.File
	tem, err = os.Create(filepath.Join(filepath.Dir(executable), ".test"))
	if tem != nil {
		_ = tem.Close()
		_ = os.Remove(tem.Name())
	}
	if os.IsPermission(err) {
		util.RunWithElevated()
		os.Exit(0)
	} else if err != nil {
		return err
	} else if !util.IsAdmin() {
		util.RunWithElevated()
		os.Exit(0)
	}

	fmt.Printf("Current version is: %s less than latest version: %s, needs to upgrade\n", cVersion, dVersion)

	var temp *os.File
	temp, err = os.CreateTemp("", "")
	if err != nil {
		err = errors.New("os.CreateTemp(\"\", \"\"): " + err.Error())
		return err
	}
	err = temp.Close()
	if err != nil {
		err = errors.New("temp.Close(): " + err.Error())
		return err
	}
	err = util.Download(client, url, temp.Name())
	if err != nil {
		err = errors.New("util.Download(client, url, temp.Name()): " + err.Error())
		return err
	}
	file, _ := os.CreateTemp("", "")
	err = file.Close()
	if err != nil {
		err = errors.New("file.Close(): " + err.Error())
		return err
	}
	err = util.UnzipKubeVPNIntoFile(temp.Name(), file.Name())
	if err != nil {
		err = errors.New("util.UnzipKubeVPNIntoFile(temp.Name(), file.Name()): " + err.Error())
		return err
	}
	err = os.Chmod(file.Name(), 0755)
	if err != nil {
		err = errors.New("os.Chmod(file.Name(), 0755): " + err.Error())
		return err
	}
	var curFolder string
	curFolder, err = os.Executable()
	if err != nil {
		err = errors.New("os.Executable(): " + err.Error())
		return err
	}
	var createTemp *os.File
	createTemp, err = os.CreateTemp("", "")
	if err != nil {
		err = errors.New("os.CreateTemp(\"\", \"\"): " + err.Error())
		return err
	}
	err = createTemp.Close()
	if err != nil {
		err = errors.New("createTemp.Close(): " + err.Error())
		return err
	}
	err = os.Remove(createTemp.Name())
	if err != nil {
		err = errors.New("os.Remove(createTemp.Name()): " + err.Error())
		return err
	}
	err = os.Rename(curFolder, createTemp.Name())
	if err != nil {
		err = errors.New("os.Rename(curFolder, createTemp.Name()): " + err.Error())
		return err
	}
	err = os.Rename(file.Name(), curFolder)
	return err
}
