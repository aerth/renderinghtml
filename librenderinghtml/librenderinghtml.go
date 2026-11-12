// Library librenderinghtml for C programs to use render_markdown (markdown2html)
//
// Be sure to use free_markdown on each output of render_markdown
package main

import (
	"os"
	"unsafe"

	"github.com/aerth/renderinghtml"
)

/*
#include <stdlib.h>
*/
import "C"

// TODO: more options etc

//export render_markdown_file
func render_markdown_file(cTitle *C.char, cFilename *C.char) *C.char {
	title := C.GoString(cTitle)
	filename := C.GoString(cFilename)
	b, err := os.ReadFile(filename)
	var result string
	if err != nil {
		result = "Error: " + err.Error()
	} else {
		result = renderinghtml.Easy(title, string(b))
	}
	return C.CString(result) // alloc
}

//export render_markdown
func render_markdown(cTitle *C.char, cContent *C.char) *C.char {
	title := C.GoString(cTitle)
	content := C.GoString(cContent)
	result := renderinghtml.Easy(title, content)
	return C.CString(result) // alloc
}

//export free_markdown
func free_markdown(str *C.char) {
	C.free(unsafe.Pointer(str))
}
func main() { panic("do not run") }
