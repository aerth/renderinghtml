package renderinghtml

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"strings"
	"text/template"

	"github.com/aerth/mostly/stackerr"
	chromahtml "github.com/alecthomas/chroma/v3/formatters/html"
	"github.com/tdewolff/minify/v2/minify"
	highlighting "github.com/yuin/goldmark-highlighting/v3"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer"
	"github.com/yuin/goldmark/v2/renderer/html"
)

var Errorf = stackerr.Errorf

//var developing = false

//go:embed "default.template.html"
var embedded embed.FS

// FS that contains "./default.template.html"
var FS fs.FS = embedded

/*
func init() {
	if developing {
		FS = os.DirFS(".")
	}
}*/

const (
	graffiti  = "oOoOoOo"
	generator = "renderinghtml v0.0.1"
)

var (
	DefaultTheme       = "tokyonight-night"
	DefaultClassPrefix = "u-" // Injected into "./default.template.html" as {{.ClassPrefix}}
)

// Cfg to modify the default before Init
func Cfg() *Options {
	if defaultoption.Renderer != nil {
		panic("already initialized")
	}
	return &defaultoption
}

var defaultoption = func() Options {
	v := Options{
		DoWrap:          false,
		CompletePage:    true,
		WithClassPrefix: "", // default
		WithCSS:         true,
		Theme:           DefaultTheme,
		Title:           "🌐",
		HeadTags:        []string{
			//
		},
		Renderer:      nil,
		Parser:        nil,
		Context:       nil,
		NoMinify:      false,
		WithClasses:   true,
		NoLineNumbers: false,
		// Context:  renderer.NewContext(),
	}
	return v
}()

func (o *Options) Init() {
	if o.Renderer != nil {
		return
	}
	o.Context = renderer.NewContext()
	/*
		t1 := time.Now()
		defer func() {
			log.Printf("done after %s", time.Since(t1).String()) // 16.063µs
		}() */
	o.Parser = parser.New(
		parser.WithExtensions(highlighting.NewParser()),
		parser.WithAutoHeadingID(),
		parser.WithAttribute(),
	)
	if o.Theme == "" {
		o.Theme = DefaultTheme
	}

	if o.WithClassPrefix == "" {
		o.WithClassPrefix = DefaultClassPrefix
	}

	o.Renderer = html.New(
		html.WithUnsafe(),
		html.WithExtensions(
			highlighting.NewHTMLRenderer(
				highlighting.WithGuessLanguage(true),

				highlighting.WithStyle(o.Theme),
				highlighting.WithFormatterOptions(
					// CSS classes
					//				chromahtml.WithModeClasses(true),
					chromahtml.WithClasses(defaultoption.WithClasses),
					chromahtml.ClassPrefix(o.WithClassPrefix),
					// LAYOUT DISABLE

					chromahtml.Standalone(false), // we add html elsewhere
					chromahtml.InlineCode(false),
					chromahtml.PreventSurroundingPre(false),
					// preferences
					chromahtml.WithLineNumbers(!defaultoption.NoLineNumbers),
					chromahtml.LineNumbersInTable(!defaultoption.NoLineNumbers),
					chromahtml.WrapLongLines(o.DoWrap),
					//				chromahtml.WithLineNumbers(true),
					chromahtml.WithModeClasses(true),
					chromahtml.WithCSSComments(false),
				),
			),
		),
	)
}

// Options has [Options.Init] method which must be called befure [Options.Markdown2HTMLString] or [Options.Markdown2HTMLString]
//
// see [defaultoption]
type Options struct {
	CompletePage    bool
	WithCSS         bool
	WithClassPrefix string
	Theme           string
	NoLineNumbers   bool
	WithClasses     bool
	DoWrap          bool // messes with line numbers
	NoMinify        bool

	html.Renderer
	parser.Parser
	Context renderer.Context
	Page
}

func (o Options) copy() Options {
	return Options{
		CompletePage:    o.CompletePage,
		WithCSS:         o.WithCSS,
		WithClassPrefix: o.WithClassPrefix,
		Theme:           o.Theme,
		DoWrap:          o.DoWrap,
		Page:            o.Page,
		NoMinify:        o.NoMinify,
		NoLineNumbers:   o.NoLineNumbers,
		WithClasses:     o.WithClasses,
		Context:         nil,
		Renderer:        nil,
		Parser:          nil,
	}
}

type Page struct {
	Title    string
	HeadTags []string
}

func New(title string) Options {
	opt := defaultoption.copy()
	opt.Title = title
	opt.Init()
	return opt
}

func Easy(title, content string) string {
	got, err := New(title).Markdown2HTMLString([]byte(content))
	if err != nil {
		println("Markdown error:", err.Error())
	}
	return got
}

func Markdown2HTMLPage(out io.Writer, in []byte, title string) error {
	if defaultoption.Renderer == nil {
		defaultoption.Init()
	}
	opt := defaultoption
	if defaultoption.Renderer == nil {
		return Errorf("no renderer")
	}
	opt.Title = title
	err := opt.Markdown2HTML(out, in)
	if err != nil {
		return err
	}
	return nil
}

func (o Options) Markdown2HTMLString(in []byte) (string, error) {
	var buf bytes.Buffer
	err := o.Markdown2HTML(&buf, in)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (o Options) Markdown2HTML(out io.Writer, in []byte) error {
	parser := o.Parser
	render := o.Renderer
	if parser == nil || render == nil {
		o.Init() // slow if repeated..
		parser = o.Parser
		render = o.Renderer
	}
	if parser == nil || render == nil {
		return Errorf("no parser/renderer")
	}

	doc := parser.Parse(in)
	cssmap := map[string]string{}
	ctx := highlighting.WithCSSMap(o.Context, cssmap)
	ext := renderer.WithContext(ctx)
	var buf bytes.Buffer
	if err := render.Render(&buf, in, doc, ext); err != nil {
		return err
	}
	if !o.CompletePage {
		_, err := buf.WriteTo(out) // DIY
		return err
	}

	// render complete page to out
	var cssBuilder strings.Builder
	for _, styleRule := range cssmap {
		cssBuilder.WriteString(styleRule)
		cssBuilder.WriteString("\n")
	}
	css, err := minify.CSS(cssBuilder.String())
	if err != nil {
		return err
	}
	if !o.WithCSS {
		css = "" // will skip style tag in [CompilePage]
	}
	if !o.CompletePage {
		if len(css) != 0 {
			_, err = out.Write([]byte(css)) // output css on first line, then the rest is the rendered html
			if err != nil {
				return err
			}
		}
		_, err = buf.WriteTo(out) // DIY
		if err != nil {
			return err
		}
		return nil
	}
	content := buf.String()
	if !o.NoMinify {
		content = must(MinifyHTML(content))
	}
	return CompilePage(out, o.Title, css, content, o.HeadTags)
}

/*
	func Render(out io.Writer, mdsrc []byte) error {
		css, content, err := Get(mdsrc, nil)
		if err != nil {
			return err
		}
		return CompletePage(out, "PHZ Server", css, content, "")
	}
*/

// CompilePage for AFTER rendering. Uses internal embedded template.
//
// Puts the title, css, (and markdown2html pre-rendered body), into a basic html page
func CompilePage(out io.Writer, title, css, content string, headtags []string) error {
	//if css == "" {
	//	_, err := out.Write([]byte(content))
	//	return err
	//}
	x := map[string]any{
		"CSS":         css + "/* " + graffiti + " */",
		"Body":        content,
		"Title":       title,
		"HeadTags":    headtags,
		"Generator":   generator,
		"ClassPrefix": DefaultClassPrefix,
	}
	return DefaultCompilePageData(out, x)
}

var DefaultCompilePageData = func(out io.Writer, x map[string]any) error {
	t, err := template.New("").
		Funcs(Funcmap).
		Option("missingkey=zero").
		ParseFS(FS, "default.template.html")
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(out, "default.template.html", x)
}

func must(s string, err error) string {
	if err != nil {
		panic(err.Error())
	}
	return s
}
