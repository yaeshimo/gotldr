package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// depend on git

// expected pages cache location:
// ~/.cache/gotldr/repo/{pages,pages.??}/PLATFORM/COMMAND.md

func getRemote(localRepo string) (string, error) {
	c := exec.Command("git", "remote", "get-url", "origin")
	c.Dir = localRepo
	b, err := c.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// clone or pull
func UpdateUpstreamPages(url string) error {
	path, err := UpstreamDir(url)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			cmd = exec.Command("git", "clone", "--depth=1", "--", url, path)
		} else {
			return err
		}
	} else {
		if fi.IsDir() {
			// check remote url
			remote, err := getRemote(path)
			if err != nil {
				return err
			}
			if url != remote {
				msg := "Specified repository is already exist and different remote url:\n"
				msg += fmt.Sprintf("\tPath  : %s\n", path)
				msg += fmt.Sprintf("\tRemote: %s\n", remote)
				return errors.New(msg)
			}
			cmd = exec.Command("git", "pull")
			cmd.Dir = path
		} else {
			return errors.New("is not directory:" + path)
		}
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// pre messages
	msg := fmt.Sprintf("directory: %s\n", cmd.Dir)
	msg += fmt.Sprintf("rmote: %s\n", url)
	msg += fmt.Sprintf("run: %q\n", cmd.Args)
	msg += "\n"
	_, err = fmt.Print(msg)
	if err != nil {
		return err
	}

	return cmd.Run()
}
