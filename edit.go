package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// for template
func sline(line string) string {
	return fmt.Sprintf("`%s`", line)
}

// TODO: consider to use text/template packages
// from: github.com/tldr-pages/tldr/contributing-guides/style-guide.md
var template = `# command-name

> Short, snappy description.
> Preferably one line; two are acceptable if neccesarry.
> More information: <https://example.com>.

- Example description:

` + sline("command -opt1 -opt2 -arg1 {{arg_value}}") + `

- Example descriptoin:

` + sline("command -opt1 -opt2") + `
`

// TODO: consider default editors
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
		// TODO: treat other platforms
	}
	for _, s := range ss {
		p, err := exec.LookPath(s)
		if err == nil {
			return p
		}
	}
	return ""
}()

// TODO: validate edited pages and print recommended messages
func Edit(name string) error {
	if editor == "" {
		return errors.New("editor not specified")
	}
	ud, err := UserPageDir()
	if err != nil {
		return err
	}
	path := filepath.Join(ud, filepath.Base(name)+".md")

	cmd := exec.Command(editor, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err := ioutil.WriteFile(path, []byte(template), 0666)
			if err != nil {
				return err
			}
			// remove new pages if not changed
			defer func() {
				b, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}
				if bytes.Equal(b, []byte(template)) {
					err := os.Remove(path)
					if err != nil {
						panic(err)
					}
				}
			}()
		} else {
			return err
		}
	} else if !fi.Mode().IsRegular() {
		return errors.New("not regular file: " + path)
	}
	return cmd.Run()
}
