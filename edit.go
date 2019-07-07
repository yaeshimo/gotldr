package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var editor = func() string {
	p := os.Getenv("EDITOR")
	if p != "" {
		return p
	}

	var ss []string
	switch runtime.GOOS {
	case "linux", "darwin", "freebsd", "netbsd", "openbsd":
		ss = []string{"vim", "emacs"}
	case "windows":
		ss = []string{"vim", "emacs", "notepad"}
	default:
		// treat other platforms?
	}
	for _, s := range ss {
		p, err := exec.LookPath(s)
		if err == nil {
			return p
		}
	}
	return ""
}()

func Edit(name string) error {
	if editor == "" {
		return errors.New("editor not specified")
	}
	ud, err := UserPageDir()
	if err != nil {
		return err
	}
	base := filepath.Base(name) + ".md"
	path := filepath.Join(ud, base)

	// make temp
	tmp, err := ioutil.TempFile("", "*_"+base)
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	var origin []byte
	origin, err = ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			origin = []byte(pageTemplate)
			// pass
		} else {
			return err
		}
	}
	_, err = tmp.Write(origin)
	if err != nil {
		return err
	}

	// edit temp
	cmd := exec.Command(editor, tmp.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		return err
	}

	// valdate
	p, err := ReadPage(tmp.Name())
	if err != nil {
		return err
	}
	if bytes.Equal(origin, p.Raw()) {
		return errors.New("page not changed")
	}

	// write page
	return ioutil.WriteFile(path, p.Raw(), 0600)
}
