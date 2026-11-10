package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aerth/mostly/flagpkg"
	"github.com/aerth/mostly/stackerr"
	"github.com/aerth/renderinghtml"
)

var (
	cfg            = renderinghtml.Cfg()
	outputfilename string //= os.Stdout.Name()
)

func Flags() {
	flag.StringVar(&renderinghtml.DefaultTheme, "theme", renderinghtml.DefaultTheme, "theme to use for code snippets")
	flag.StringVar(&renderinghtml.DefaultClassPrefix, "class-prefix", renderinghtml.DefaultClassPrefix, "to target element selectors")

	flagpkg.InverseBoolVar(&cfg.CompletePage, "no-complete", cfg.CompletePage, "dont make whole html page")
	flagpkg.InverseBoolVar(&cfg.WithCSS, "no-css", cfg.WithCSS, "dont make css")
	flag.StringVar(&cfg.Title, "title", cfg.Title, "title element")

	flag.StringVar(&outputfilename, "o", outputfilename, "output to stdout or file")
}

func main() {
	Flags()
	sortargs(os.Args[1:])
	flag.Parse()
	if outputfilename == "-" {
		outputfilename = os.Stdout.Name()
	}
	err := main2(outputfilename, flag.Args())
	if err != nil {
		log.Fatalf("Fatal: %+v", err)
	}
}

func sortargs(s []string) {
	at := -1
	for i := range s {
		if strings.HasPrefix(s[i], "-") {
			at = i
			break
		}
	}
	if at > 0 {
		x := append(s[at:], s[0:at]...)
		copy(s, x)
	}
}

func main2(outputfilename string, filenames []string) error {
	if len(filenames) == 0 {
		filenames = []string{"-"}
	}
	for _, filename := range filenames {
		if err := main3(outputfilename, filename); err != nil {
			return err
		}
	}
	return nil
}

func readfile(name string) ([]byte, error) {
	if name == "-" || name == "stdin" {
		name = os.Stdin.Name()
	}
	return os.ReadFile(name)
}

func main3(outputfilename string, filename string) error {
	mdbytes, err := readfile(filename)
	if err != nil {
		return stackerr.Errorf("readfile: %w", err)
	}
	filename = strings.TrimSuffix(filename, ".md") + ".html"

	var tempfile *os.File
	switch outputfilename {
	case "": // auto name in /tmp
		t := filepath.Join(os.TempDir(), "markdown2html", filename)
		_ = os.MkdirAll(filepath.Dir(t), 0o755)
		tempfile, err = os.Create(t)
		if err != nil {
			return stackerr.Errorf("createTemp: %w", err)
		}
	case "-", os.Stdout.Name(): // no name
		tempfile = os.Stdout
	default: // all to one output name
		tempfile, err = os.OpenFile(outputfilename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0o644)
		if err != nil {
			return stackerr.Errorf("open output: %w", err)
		}
	}
	// out
	if err := cfg.Markdown2HTML(tempfile, mdbytes); err != nil {
		return stackerr.Errorf("markdown2html: %w", err)
	}
	if tempfile != os.Stdout {
		if e := tempfile.Close(); e != nil {
			slog.Error("closing output file", "err", e, "name", tempfile.Name())
		}
	}
	if outputfilename == "" {
		fmt.Println(tempfile.Name())
	}
	return nil
}
