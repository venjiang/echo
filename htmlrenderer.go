package echo

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// template
type HtmlRender struct {
	rw        http.ResponseWriter
	templates *template.Template
	opt       Options
}

func HtmlRenderer(options ...Options) Renderer {
	funcMap := template.FuncMap{
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
		"js": func(s string) string {
			return template.JSEscapeString(s)
		},
	}

	opt := prepareOptions(options)
	opt.Funcs = []template.FuncMap{funcMap,}
	t := compileTemplates(opt)
	return &HtmlRender{templates: t, opt: opt}
}

// Render HTML
func (r *HtmlRender) Render(wr io.Writer, name string, data interface{}) error {
	// assign a layout if there is one
	if len(r.opt.Layout) > 0 {
		r.addLayoutFuncs(name, data)
		name = r.opt.Layout
	}

	err := r.execute(wr, name, data)
	if err != nil {
		//		http.Error(r, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

// Delims represents a set of Left and Right delimiters for HTML template rendering
type Delims struct {
	// Left delimiter, defaults to {{
	Left string
	// Right delimiter, defaults to }}
	Right string
}

// Included helper functions for use when rendering html
var helperFuncs = template.FuncMap{
	"yield": func() (string, error) {
		return "", fmt.Errorf("yield called with no layout defined")
	},
	"current": func() (string, error) {
		return "", nil
	},
	"block": func(blockName string, required bool) (string, error) {
		return "", fmt.Errorf("block called with no layout defined")
	},
}

// Options is a struct for specifying configuration options for the render.Renderer middleware
type Options struct {
	// Directory to load templates. Default is "templates"
	Directory string
	// Layout template name. Will not render a layout if "". Defaults to "".
	Layout string
	// Extensions to parse template files from. Defaults to [".tmpl"]
	Extensions []string
	// Funcs is a slice of FuncMaps to apply to the template upon compilation. This is useful for helper functions. Defaults to [].
	Funcs []template.FuncMap
	// Delims sets the action delimiters to the specified strings in the Delims struct.
	Delims Delims
}

func prepareOptions(options []Options) Options {
	var opt Options
	if len(options) > 0 {
		opt = options[0]
	}

	// Defaults
	if len(opt.Directory) == 0 {
		opt.Directory = "templates"
	}
	if len(opt.Extensions) == 0 {
		opt.Extensions = []string{".html"}
	}

	return opt
}

func compileTemplates(options Options) *template.Template {
	dir := options.Directory
	t := template.New(dir)
	t.Delims(options.Delims.Left, options.Delims.Right)
	// parse an initial template in case we don't have any
	template.Must(t.Parse("root"))

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		r, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		ext := getExt(r)

		for _, extension := range options.Extensions {
			if ext == extension {

				buf, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}

				//				name := (r[0 : len(r) - len(ext)])
				name := r
				tmpl := t.New(filepath.ToSlash(name))

				// add our funcmaps
				for _, funcs := range options.Funcs {
					tmpl.Funcs(funcs)
				}

				// Bomb out if parse fails. We don't want any silent server starts.
				template.Must(tmpl.Funcs(helperFuncs).Parse(string(buf)))
				break
			}
		}

		return nil
	})

	return t
}

func (r *HtmlRender) execute(wr io.Writer, name string, binding interface{}) error {
	return r.templates.ExecuteTemplate(wr, name, binding)
}

func (r *HtmlRender) addLayoutFuncs(name string, binding interface{}) {
	funcs := template.FuncMap{
		"yield": func() (template.HTML, error) {
			buf := new(bytes.Buffer)
			err := r.execute(buf, name, binding)
			return template.HTML(buf.String()), err
		},
		"current": func() (string, error) {
			return name, nil
		},
		"block": func(blockName string) (template.HTML, error) {
			if r.templates.Lookup(blockName) != nil {
				buf := new(bytes.Buffer)
				err := r.execute(buf, blockName, binding)
				return template.HTML(buf.String()), err
			}
			return "", nil
		},
	}
	r.templates.Funcs(funcs)
}

func getExt(s string) string {
	if strings.Index(s, ".") == -1 {
		return ""
	}
	return "." + strings.Join(strings.Split(s, ".")[1:], ".")
}
