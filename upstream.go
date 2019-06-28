package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// depend on git

// expected pages cache location:
// upstreamCacheDir/{pages,pages.??}/PLATFORM/COMMAND.md

var upstream = "https://github.com/tldr-pages/tldr.git"

func UpstreamCacheDir() (string, error) {
	ud, err := CacheHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(ud, "tldr"), nil
}

// clone or pull
func UpdateUpstreamPages() error {
	path, err := UpstreamCacheDir()
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			cmd = exec.Command("git", "clone", "--depth=1", "--", upstream, path)
		} else {
			return err
		}
	} else {
		if fi.IsDir() {
			cmd = exec.Command("git", "pull", "--depth=1")
			cmd.Dir = path
		} else {
			return errors.New("is not directory:" + path)
		}
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// pre messages
	msg := fmt.Sprintf("Work directory: %q\n", cmd.Dir)
	msg += fmt.Sprintf("Run: %q\n", cmd.Args)
	msg += "\n"
	_, err = fmt.Print(msg)
	if err != nil {
		return err
	}

	return cmd.Run()
}
