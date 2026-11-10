package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/aerth/renderinghtml"
)

var cfgFS fs.FS = nil

var serveReadme = false

func main() {
	cfg := renderinghtml.Cfg()
	var addr string
	var directory string
	cfg.HeadTags = []string{
		`<link rel="icon" href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==">`,
	}

	flag.BoolVar(&serveReadme, "serve-readme", serveReadme, "combined with -d flag, serve README.md instead of index.md")
	flag.BoolFunc("no-minify", "disable minify", func(s string) error {
		renderinghtml.Cfg().NoMinify = s != "false"
		return nil
	})
	flag.StringVar(&directory, "d", directory, "if nonempty, serve directory as html (fileserver*)")
	flag.StringVar(&addr, "addr", addr, "addr")
	flag.Func("meta", "add meta tag to head (can be repeated)", func(s string) error {
		s = strings.TrimSpace(s)
		if !strings.HasPrefix(s, "<") || !strings.HasSuffix(s, ">") {
			return fmt.Errorf("need html tag for head, such as <style></style>")
		}
		cfg.HeadTags = append(cfg.HeadTags, s)
		return nil
	})
	flag.Parse()
	if flag.NArg() != 0 {
		println("unknown arguments")
		return
	}
	if directory != "" {
		{
			readmemode := "README.md"
			if !serveReadme {
				readmemode = "index.md"
			}
			slog.Info("serving dir", "path", directory, "addr", addr, "mode", readmemode)
		}
		f, err := fs.Sub(os.DirFS(directory), directory)
		if err != nil {
			panic(err.Error())
		}
		cfgFS = f
	} else {
		slog.Info("serving dummy page", "addr", addr)
	}
	mux := http.NewServeMux()
	if directory != "" {
		mux.HandleFunc("GET /", serveFS)
	} else {
		mux.HandleFunc("GET /{$}", serveDummyPage)
	}
	log.Fatalln(http.ListenAndServe(addr, mux))
}

func serveFS(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Path
	if s == "" || s[0] != '/' {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if strings.HasSuffix(s, "/") {
		s += "index.html"
	}
	s = s[1:]

	if strings.HasSuffix(s, ".md") {
		w.WriteHeader(404)
		return
	}
	if before, ok := strings.CutSuffix(s, ".html"); ok {
		s = before + ".md"
	}
	if !strings.HasSuffix(s, ".md") {
		slog.Info("serving file", "name", s)

		http.ServeFileFS(w, r, cfgFS, s)
		return
	}
	if serveReadme && s == "index.md" {
		s = "README.md"
	}
	title := "TODO"
	slog.Info("serving file", "name", s)
	md, err := fs.ReadFile(cfgFS, s)
	if err != nil {
		slog.Error("readfile", "err", err, "path", s)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	out, err := renderinghtml.New(title).Markdown2HTMLString(md)
	if err != nil {
		slog.Error("render", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	x, err := renderinghtml.MinifyHTML(out)
	if err != nil {
		slog.Error("cant minify", "err", err)
		x = out
	}
	_, err = w.Write([]byte(x))
	if err != nil {
		slog.Error("writing", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func serveDummyPage(w http.ResponseWriter, r *http.Request) {
	title := "It works! 👷"
	md := []byte("# It works!\n```go\nvar k = 4\nprintln(k)\n```\n<p>It does</p><hr><h3>Under Construction</h3>")
	buf := bytes.Buffer{}
	err := renderinghtml.Markdown2HTMLPage(&buf, md, title)
	if err != nil {
		slog.Error("render", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = buf.WriteTo(w)
	if err != nil {
		slog.Error("writeconn", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
