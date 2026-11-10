package renderinghtml

import (
	"os"
	"testing"
)

func TestRender1(t *testing.T) {
	x := "# wow cool\n\n```go\nvar x = 42\n```\n"
	op := Options{}
	op.Init()
	op.CompletePage = true
	op.WithCSS = true
	err := op.Markdown2HTML(os.Stdout, []byte(x))
	if err != nil {
		t.Fatal(err)
	}
}
