package wailsselfupdate

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	goruntime "runtime"

	"github.com/blang/semver"
	gselfupdate "github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	ErrBadUrl = errors.New("bad url")
	ErrWrite  = errors.New("write error")
)

var _ Updater = &updater{}

type Updater interface {
	CheckForUpdates() error
	CheckForUpdatesSilent() error
	CheckForUpdatesBackground() error
	HasUpdate() (*gselfupdate.Release, bool, error)
}

type updater struct {
	currentVersion semver.Version
	wailsContext   context.Context
	slug           string
}

func NewUpdater(currentVersion semver.Version, wailsContext context.Context, slug string) *updater {
	return &updater{
		currentVersion: currentVersion,
		wailsContext:   wailsContext,
		slug:           slug,
	}
}

func (u *updater) CheckForUpdatesBackground() error {
	release, found, err := u.HasUpdate()

	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	// Update available
	response := u.displayUpdateDialog(release.Version, release.ReleaseNotes)

	if response == "Update" || response == "Yes" {
		if goruntime.GOOS == "darwin" {
			u.updateDarwin(release)
		} else if goruntime.GOOS == "windows" {
			u.updateWindows(release)
		} else {
			return errors.New("unsupported")
		}
	}

	return nil
}

func (u *updater) CheckForUpdates() error {
	release, found, err := u.HasUpdate()

	if err != nil {
		return err
	}

	if !found {
		u.displayNoUpdatesAvailable()
		return nil
	}

	// Update available
	response := u.displayUpdateDialog(release.Version, release.ReleaseNotes)

	if response == "Update" || response == "Yes" {
		if goruntime.GOOS == "darwin" {
			u.updateDarwin(release)
		} else if goruntime.GOOS == "windows" {
			u.updateWindows(release)
		} else {
			return errors.New("unsupported")
		}
	}

	return nil
}

func (u *updater) CheckForUpdatesSilent() error {
	release, found, err := u.HasUpdate()

	if err != nil {
		log.Println("failed to detect for updates: ", err)
		return err
	}

	if !found {
		log.Println("no updates available")
		return nil
	}

	log.Printf("update found %s --> %s\n", u.currentVersion.String(), release.Version.String())

	if goruntime.GOOS == "darwin" {
		log.Println("updating for darwin...")
		u.updateDarwin(release)
	} else if goruntime.GOOS == "windows" {
		log.Println("updating for windows...")
		u.updateWindows(release)
	} else {
		log.Println("os not supported...")
		return errors.New("unsupported")
	}

	return nil
}

func (u *updater) HasUpdate() (*gselfupdate.Release, bool, error) {
	release, found, err := gselfupdate.DetectLatest(u.slug)

	if err != nil {
		return nil, false, err
	}

	if !found || release.Version.LTE(u.currentVersion) {
		return nil, false, nil
	}

	return release, true, nil

}

func (u *updater) download(url string, path string) error {
	res, err := http.Get(url)

	if err != nil {
		return ErrBadUrl
	}

	out, err := os.Create(path)

	if err != nil {
		return ErrWrite
	}

	if _, err := io.Copy(out, res.Body); err != nil {
		return ErrWrite
	}

	return nil
}

func (u *updater) getDownloadsDir() (string, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	return path.Join(homeDir, "Downloads"), nil
}

func (u *updater) displayNoUpdatesAvailable() {
	runtime.MessageDialog(u.wailsContext, runtime.MessageDialogOptions{
		Type:          runtime.InfoDialog,
		Title:         "No Updates Available",
		Message:       "Current version: " + u.currentVersion.String(),
		Buttons:       []string{"OK"},
		DefaultButton: "OK",
	})
}

func (u *updater) displayUpdateDialog(newVersion semver.Version, releaseNotes string) string {
	message := "Version " + newVersion.String() + " is available\n\n Release notes:\n" + releaseNotes

	opts := runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "Update Available",
		Message:       message,
		Buttons:       []string{"Update", "Cancel"},
		DefaultButton: "Update",
		CancelButton:  "Cancel",
	}

	dialog, _ := runtime.MessageDialog(u.wailsContext, opts)

	return dialog

}
