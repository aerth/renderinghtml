package renderinghtml

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"text/template"

	mhtml "github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/minify"
)

var Funcmap = template.FuncMap{
	"minifyCSS":   MinifyCSS,
	"minifyJS":    MinifyJS,
	"minifySVG":   MinifySVG,
	"minifyHTML":  MinifyHTML,
	"minifyXML":   MinifyXML,
	"minifyJSON":  MinifyJSON,
	"minifyHTML2": MinifyHTML2,
}

// minifiers
var (
	MinifyCSS  = minify.CSS
	MinifyJS   = minify.JS
	MinifySVG  = minify.SVG
	MinifyHTML = minify.HTML
	MinifyXML  = minify.XML
	MinifyJSON = minify.JSON
)

var Minify = minify.Default

func MinifyHTML2(s string) (string, error) {
	var buf bytes.Buffer
	err := minifyHTML2(&buf, strings.NewReader(s), nil)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

var HTMLMinifier = mhtml.Minifier{
	KeepDefaultAttrVals: true,
	KeepDocumentTags:    true,
	KeepQuotes:          true,
	KeepWhitespace:      true,
	KeepEndTags:         true,
}

func minifyHTML2(w io.Writer, r io.Reader, params map[string]string) error {
	m := Minify
	return new(HTMLMinifier).Minify(m, w, r, params)
}

func Minifier(ct string, in string) string {
	x, err := Minify.String(ct, in)
	if err != nil {
		slog.Error("minify", "ct", ct, "err", err)
		return ""
	}
	return x
}

func MinifierStream(mediatype string, out io.Writer, in io.Reader) (n int64, err error) {
	return io.Copy(out, Minify.Reader(mediatype, in))
}
