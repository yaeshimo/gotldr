package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

type Example struct {
	Desc string
	Line string
}

// too long don't read
type Page struct {
	Path string

	Name     string
	Descs    []string
	Examples []*Example
}

// TODO: fix print style
// add color
// parse for {{var}}
// trim "$" for gotldr vim, emacs, etc...
func (p *Page) String() string {
	s := fmt.Sprintf("Usage of %q. Location: %s\n\n", p.Name, p.Path)

	s += "Description:\n"
	for _, desc := range p.Descs {
		s += fmt.Sprintf("\t%s\n", desc)
	}
	s += "\n"

	s += "Examples:\n"
	for _, eg := range p.Examples {
		s += fmt.Sprintf("\t%s\n", eg.Desc)
		s += fmt.Sprintf("\t$ %s\n", eg.Line)
		s += "\n"
	}
	return s
}

// TODO: add tests
func LazyParsePage(b []byte) (*Page, error) {
	page := new(Page)
	ss := strings.SplitN(string(bytes.TrimSpace(b)), "\n\n", 3)
	if len(ss) != 3 {
		return nil, errors.New("can not split to 3 blocks: Name, Desc and Examples")
	}

	// page.Name
	if !strings.HasPrefix(ss[0], "# ") {
		return nil, errors.New("not found prefix \"# \"")
	}
	page.Name = strings.TrimSpace(strings.TrimPrefix(ss[0], "# "))

	// page.Descs
	for _, s := range strings.Split(ss[1], "\n") {
		if !strings.HasPrefix(s, "> ") {
			return nil, errors.New("not foudn prefix \"> \"")
		}
		s = strings.TrimSpace(strings.TrimPrefix(s, "> "))
		page.Descs = append(page.Descs, s)
	}

	// page.Examples
	sc := bufio.NewScanner(strings.NewReader(ss[2]))
	egn := 0
	// for - Desc...
	for sc.Scan() {
		text := strings.TrimSpace(sc.Text())
		if text == "" {
			continue
		}
		if !strings.HasPrefix(text, "- ") || !strings.HasSuffix(text, ":") {
			return nil, errors.New("invalid command descriptin")
		}
		text = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(text, "- "), ":"))
		eg := &Example{Desc: text}
		egn++
		// for `cmd arg...`
		for sc.Scan() {
			text := strings.TrimSpace(sc.Text())
			if text == "" {
				continue
			}
			if !strings.HasPrefix(text, "`") || !strings.HasSuffix(text, "`") {
				return nil, errors.New("invalid example of command lines")
			}
			eg.Line = strings.TrimSuffix(strings.TrimPrefix(text, "`"), "`")
			egn++
			page.Examples = append(page.Examples, eg)
			break
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if egn%2 != 0 {
		return nil, errors.New("not enough examples")
	}

	return page, nil
}

func ReadPage(path string) (*Page, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	page, err := LazyParsePage(b)
	if err != nil {
		return nil, fmt.Errorf("%s: %q", err.Error(), path)
	}
	page.Path = path
	return page, nil
}
