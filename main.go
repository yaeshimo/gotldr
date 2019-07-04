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

	"github.com/mattn/go-colorable"
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

// TODO: add -list for list tldr pages

// TODO: add tests
// return first founded path to pages

// TODO: add -rm for remove users pages

// TODO: add -add for add examples to users pages
// e.g. gotldr -add 'Command Arg1 Arg2' 'Example description'

const DefaultUpstream = "https://github.com/tldr-pages/tldr.git"

var usage = `Usage:
  gotldr [Options] [COMMAND]

Options:
  -help                  Display this message
  -e, -edit COMMAND      Edit your tldr pages with $EDITOR
  -p, -platform PLATFORM Specify target platforms
  -l, -lang LANG         Specify priority languages with ISO 639-1 codes
  -r, -remote URL        Specify upstream URL
                         (default "` + DefaultUpstream + `")
  -u, -update            Update or download tldr pages into local
                         from -remote URL with git
  -nocolor               Disable color output
  -dirs                  Display current candidate directories with index
  -index INDEX           Specify pages directory by index

Examples:
  $ gotldr -help        # help
  $ gotldr cd           # show usage of cd
  $ gotldr -edit gotldr # edit or make your own tldr pages

  # change the upstream
  $ alias='gotldr -remote="https://gitlab.com/USER/REPO.git"'
  $ gotldr -update
`

var usageWriter io.Writer = os.Stderr

func printUsage() { fmt.Fprintln(usageWriter, usage) }

var opt struct {
	help     bool
	edit     string
	platform string
	lang     string
	remote   string
	update   bool
	nocolor  bool
	dirs     bool
	index    int
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
	flag.StringVar(&opt.edit, "edit", "", "Edit users tldr pages")
	flag.StringVar(&opt.edit, "e", "", "Alias of -edit")
	flag.StringVar(&opt.platform, "platform", defp, "Set target platforms")
	flag.StringVar(&opt.platform, "p", defp, "Alias of -platform")
	flag.StringVar(&opt.lang, "lang", deflang, "Set target language with ISO 639-1 codes")
	flag.StringVar(&opt.lang, "l", deflang, "Alias of -lang")
	flag.StringVar(&opt.remote, "remote", DefaultUpstream, `Specify upstream URL`)
	flag.StringVar(&opt.remote, "r", DefaultUpstream, "Alias of -remote")
	flag.BoolVar(&opt.update, "update", false, "Update pages")
	flag.BoolVar(&opt.update, "u", false, "Alias of -update")
	flag.BoolVar(&opt.nocolor, "nocolor", false, "Disable color output")
	flag.BoolVar(&opt.dirs, "dirs", false, "Display candidate directories")
	flag.IntVar(&opt.index, "index", -1, "Specify pages directory with index")

	flag.Usage = printUsage
}

func run() error {
	flag.Parse()
	if opt.help {
		usageWriter = os.Stdout
		flag.Usage()
		return nil
	}

	if opt.update {
		if flag.NArg() != 0 {
			return errors.New("unexpected arguments: " + strings.Join(flag.Args()[1:], " "))
		}
		return UpdateUpstreamPages(opt.remote)
	}

	if opt.edit != "" {
		if flag.NFlag() != 1 {
			return errors.New("too many specified flags")
		}
		if flag.NArg() != 0 {
			return errors.New("unexpected arguments: " + strings.Join(flag.Args(), " "))
		}
		return Edit(filepath.Base(opt.edit))
	}

	dirs, err := CandidateCacheDirs(opt.remote, opt.platform, opt.lang)
	if err != nil {
		return err
	}
	if opt.index >= 0 {
		if opt.index >= len(dirs) {
			return errors.New("index out of bounds")
		}
		dirs = []string{dirs[opt.index]}
	}

	if opt.dirs {
		msg := fmt.Sprintf("[Index]\t[Directory]\n")
		for i, d := range dirs {
			msg += fmt.Sprintf("%4d\t%q\n", i, d)
		}
		_, err = fmt.Println(msg)
		return err
	}

	// read page
	switch flag.NArg() {
	case 0:
		return errors.New("command name not specified")
	case 1:
		// pass
	default:
		return errors.New("unexpected arguments: " + strings.Join(flag.Args(), " "))
	}
	path, err := FindPage(dirs, flag.Arg(0))
	if err != nil {
		return err
	}
	page, err := ReadPage(path)
	if err != nil {
		return err
	}

	var stdout io.Writer = os.Stdout
	if opt.nocolor {
		stdout = colorable.NewNonColorable(stdout)
	} else {
		stdout = colorable.NewColorableStdout()
	}

	_, err = fmt.Fprint(stdout, page.Wrap())
	return err
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
