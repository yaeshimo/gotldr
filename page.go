package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

// for pageTemplate
func sline(line string) string {
	return fmt.Sprintf("`%s`", line)
}

// TODO: consider to use text/template packages
// References: https://github.com/tldr-pages/tldr/contributing-guides/style-guide.md
var pageTemplate = `# command-name

> Short, snappy description.
> Preferably one line; two are acceptable if neccesarry.
> More information: <https://example.com>.

- Example description:

` + sline("command -opt1 -opt2 -arg1 {{arg_value}}") + `

- Example descriptoin:

` + sline("command -opt1 -opt2") + `
`

const (
	reset = "\x1b[0m"
	bold  = "\x1b[1m"
	under = "\x1b[4m"
	red   = "\x1b[31m"
	green = "\x1b[32m"
)

type Example struct {
	Desc string
	Line string
}

type Page struct {
	Path string

	Name     string
	Descs    []string
	Examples []*Example
}

var matchVariable = regexp.MustCompile("{{([^(}})]*)}}")

// wrap with ANSI colors
func (p *Page) Wrap() *Page {
	examples := make([]*Example, len(p.Examples))
	for i, eg := range p.Examples {
		examples[i] = &Example{
			Desc: green + eg.Desc + reset,
			Line: matchVariable.ReplaceAllString(eg.Line, bold+"$1"+reset),
		}
	}
	return &Page{
		Path:     p.Path,
		Name:     p.Name,
		Descs:    p.Descs,
		Examples: examples,
	}
}

func (p *Page) String() string {
	s := fmt.Sprintf("Usage of %s (Location: %s).\n\n", p.Name, p.Path)

	s += "Description:\n"
	for _, desc := range p.Descs {
		s += fmt.Sprintf("\t%s\n", desc)
	}
	s += "\n"

	s += "Examples:\n"
	for _, eg := range p.Examples {
		s += fmt.Sprintf("\t%s\n", eg.Desc)
		s += fmt.Sprintf("\t  %s\n", eg.Line)
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
