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
	path string
	raw  []byte

	Name     string
	Descs    []string
	Examples []*Example
}

func (p *Page) Raw() []byte { return p.raw }

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
		path:     p.path,
		Name:     p.Name,
		Descs:    p.Descs,
		Examples: examples,
	}
}

func (p *Page) String() string {
	s := fmt.Sprintf("Usage of %s:\n", p.Name)
	s += fmt.Sprintf("\tLocation: %s\n\n", p.path)

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

type ParseError struct {
	path string
	err  error
}

func (pe *ParseError) Error() string {
	return fmt.Sprintf("parse error:%s:%s", pe.err.Error(), pe.path)
}

// TODO: add tests
func lazyParse(b []byte) (*Page, error) {
	page := &Page{raw: b}
	ss := strings.SplitN(string(bytes.TrimSpace(b)), "\n\n", 3)
	if len(ss) != 3 {
		return nil, errors.New("can not split to 3 contexts")
	}

	// page.Name
	if !strings.HasPrefix(ss[0], "# ") {
		return nil, errors.New("not found prefix \"# \"")
	}
	page.Name = strings.TrimSpace(strings.TrimPrefix(ss[0], "# "))

	// page.Descs
	for _, s := range strings.Split(ss[1], "\n") {
		if !strings.HasPrefix(s, "> ") {
			return nil, errors.New("not found prefix \"> \"")
		}
		s = strings.TrimSpace(strings.TrimPrefix(s, "> "))
		page.Descs = append(page.Descs, s)
	}

	// page.Examples
	sc := bufio.NewScanner(strings.NewReader(ss[2]))
	var eg *Example
	// for - Desc...
	for sc.Scan() {
		text := strings.TrimSpace(sc.Text())
		if text == "" {
			continue
		}
		if !strings.HasPrefix(text, "- ") || !strings.HasSuffix(text, ":") {
			return nil, errors.New("invalid command descriptins")
		}
		text = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(text, "- "), ":"))
		eg = &Example{Desc: text}
		// for `cmd arg...`
		for sc.Scan() {
			text := strings.TrimSpace(sc.Text())
			if text == "" {
				continue
			}
			if !strings.HasPrefix(text, "`") || !strings.HasSuffix(text, "`") {
				return nil, errors.New("invalid examples")
			}
			eg.Line = strings.TrimSuffix(strings.TrimPrefix(text, "`"), "`")
			page.Examples = append(page.Examples, eg)
			break
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if eg != nil && eg.Line == "" {
		return nil, errors.New("not found command examples")
	}

	return page, nil
}

func ReadPage(path string) (*Page, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	page, err := lazyParse(b)
	if err != nil {
		return nil, &ParseError{path: path, err: err}
	}
	page.path = path
	return page, nil
}
