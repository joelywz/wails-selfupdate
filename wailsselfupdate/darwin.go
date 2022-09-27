package wailsselfupdate

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	gselfupdate "github.com/rhysd/go-github-selfupdate/selfupdate"
)

func (u *updater) replaceDarwinApp(archivePath string, appPath string) error {
	cmd := exec.Command("ditto", "-xk", archivePath, appPath)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (u *updater) getDarwinAppPath() (string, error) {
	appPath, err := os.Executable()

	if err != nil {
		return "", err
	}

	appPath = filepath.Join(appPath, "..", "..", "..", "..")
	return appPath, nil
}

func (u *updater) updateDarwin(release *gselfupdate.Release) error {

	var downloadsDir string
	var err error

	if downloadsDir, err = u.getDownloadsDir(); err != nil {
		return err
	}

	downloadPath := path.Join(downloadsDir, fmt.Sprintf("%s-%s.zip", release.Name, release.Version))

	if err := u.download(release.AssetURL, downloadPath); err != nil {
		return err
	}

	var appPath string

	if appPath, err = u.getDarwinAppPath(); err != nil {
		return err
	}

	if err := u.replaceDarwinApp(downloadPath, appPath); err != nil {
		return err
	}

	return nil

}
