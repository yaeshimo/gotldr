package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Done:
// show tldr
// 1. make candidatedirs with PLATFORM, (LANGUAGES or LANG)
// 2. search command pages from candidates
// 3. output tldr

// Done:
// update tldr
// 1. git clone from upstream to cache directory
// note: that expected CACHEDIR/tldr/HOST/REPO/{pages,pages.??}/COMMAND.md

// Done:
// edit tldr for users
// note: pages location is CACHEDIR/tldr/user/{pages,pages.??}/COMMAND.md
// respect $EDITOR
// 1. if exist specify pages then to edit
// 2. if not exist then make template
// 3. if template not modified then not save
// 4. if template modified then save with verification

// TODO: add color for *Pages.String()

// TODO: add -list flags for list tldr pages

// TODO: add tests
// return first founded path to pages

const usage = `Usage:
  gotldr [Options] COMMAND

Options:
  -help                    Display this message
  -e, -edit [COMMAND]      Edit your tldr pages with $EDITOR
  -p, -platform [PLATFORM] Specify target platforms
  -l, -lang [LANG]         Specify priority languages with ISO 639-1 codes
  -u, -update              Update or download tldr pages (depend on git)

Examples:
  $ gotldr -help       # help
  $ gotldr tar         # show usage of tar
  $ gotldr -edit rsync # edit or make your tldr pages
`

var usageWriter io.Writer = os.Stderr

func printUsage() { fmt.Fprintln(usageWriter, usage) }

var opt struct {
	help     bool
	edit     bool
	platform string
	lang     string
	update   bool
}

func init() {
	// TODO: fix deflang
	var (
		defp    string
		deflang string
	)
	switch n := runtime.GOOS; n {
	case "linux":
		defp = n
		deflang = os.Getenv("LANG")
	case "windows":
		defp = n
	case "darwin":
		defp = "osx"
		deflang = os.Getenv("LANG")
	case "solaris":
		defp = "sunos"
	}
	if len(deflang) > 2 {
		deflang = deflang[:2]
		if !IsValidLang(deflang) {
			deflang = ""
		}
	} else {
		deflang = ""
	}

	flag.BoolVar(&opt.help, "help", false, "Display this message")
	flag.BoolVar(&opt.edit, "edit", false, "Edit users tldr pages")
	flag.BoolVar(&opt.edit, "e", false, "Alias of -edit")
	flag.StringVar(&opt.platform, "platform", defp, "Set target platforms")
	flag.StringVar(&opt.platform, "p", defp, "Alias of -platform")
	flag.StringVar(&opt.lang, "lang", deflang, "Set target language with ISO 639-1 codes")
	flag.StringVar(&opt.lang, "l", deflang, "Alias of -lang")
	flag.BoolVar(&opt.update, "update", false, "Update pages")
	flag.BoolVar(&opt.update, "u", false, "Alias of -update")

	flag.Usage = printUsage
}

func run() error {
	flag.Parse()
	if opt.help {
		usageWriter = os.Stdout
		flag.Usage()
		return nil
	}
	switch flag.NArg() {
	case 1:
		// pass
	case 0:
		if opt.update {
			if flag.NFlag() != 1 {
				flag.Usage()
				return errors.New("too many specified flags")
			}
			return UpdateUpstreamPages()
		}
		flag.Usage()
		return errors.New("command not specified")
	default:
		flag.Usage()
		return errors.New("unexpected arguments: " + strings.Join(flag.Args(), " "))
	}

	if opt.edit {
		if flag.NFlag() != 1 {
			flag.Usage()
			return errors.New("too many specified flags")
		}
		return Edit(filepath.Base(flag.Arg(0)))
	}

	dirs, err := CandidateCacheDirs(opt.platform, opt.lang)
	if err != nil {
		return err
	}
	path, err := FindPage(dirs, flag.Arg(0))
	if err != nil {
		return err
	}
	page, err := ReadPage(path)
	if err != nil {
		return err
	}
	_, err = fmt.Print(page)
	return err
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
