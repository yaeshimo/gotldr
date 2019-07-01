package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// depend on git

// expected pages cache location:
// ~/.cache/gotldr/repo/{pages,pages.??}/PLATFORM/COMMAND.md

const DefaultUpstream = "https://github.com/tldr-pages/tldr.git"

var (
	mux      sync.Mutex
	upstream = DefaultUpstream
)

func SetUpstream(url string) {
	mux.Lock()
	upstream = url
	mux.Unlock()
}
func Upstream() string {
	mux.Lock()
	defer mux.Unlock()
	return upstream
}

// expected location: ~/.cache/gotldr/repo/REPO/
func UpstreamLocalRepo() (string, error) {
	path, err := CacheHome()
	if err != nil {
		return "", err
	}
	path = filepath.Join(path, "repo")
	err = os.MkdirAll(path, 0700)
	if err != nil {
		return "", err
	}
	base := strings.TrimSuffix(filepath.Base(Upstream()), ".git")
	return filepath.Join(path, base), nil
}

func getRemote(repo string) (string, error) {
	c := exec.Command("git", "remote", "get-url", "origin")
	c.Dir = repo
	b, err := c.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// clone or pull
func UpdateUpstreamPages() error {
	path, err := UpstreamLocalRepo()
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
			// check remote url
			remote, err := getRemote(path)
			if err != nil {
				return err
			}
			if Upstream() != remote {
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
	msg := fmt.Sprintf("Work directory: %q\n", cmd.Dir)
	msg += fmt.Sprintf("Run: %q\n", cmd.Args)
	msg += "\n"
	_, err = fmt.Print(msg)
	if err != nil {
		return err
	}

	return cmd.Run()
}
