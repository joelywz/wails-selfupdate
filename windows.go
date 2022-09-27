package wailsselfupdate

import (
	"os"

	gselfupdate "github.com/rhysd/go-github-selfupdate/selfupdate"
)

func (u *updater) updateWindows(release *gselfupdate.Release) error {
	appPath, err := os.Executable()

	if err != nil {
		return err
	}

	return gselfupdate.UpdateTo(release.AssetURL, appPath)
}
