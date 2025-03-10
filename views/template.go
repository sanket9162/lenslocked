package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/sanket9162/lenslocked/context"
	"github.com/sanket9162/lenslocked/models"
)
type Template struct {
	htmlTpl *template.Template
}


func Must(t Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return t 
}

func ParseFS(fs fs.FS, pattern ...string) (Template, error) {
	tpl := template.New(pattern[0])
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfFeild": func() (template.HTML, error) {
				return "", fmt.Errorf("csrfField not implemented")
			},
			"currentUser": func() (template.HTML, error) {
				return "", fmt.Errorf("csrfField not implemented")
			},
		},
	)
	tpl, err := tpl.ParseFS(fs, pattern...)
	 if err != nil{
		return Template{}, fmt.Errorf("parsing template: %w", err)
	}

	return Template{
		htmlTpl:tpl,
	}, nil
}

func (t Template) Execute (w http.ResponseWriter, r *http.Request, data interface{}){
	tpl, err := t.htmlTpl.Clone()
	if err != nil {
		log.Printf("Cloning template: %v", err)
		http.Error(w, "There was an error rendering the page.", http.StatusInternalServerError)
		return
	}
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfFeild": func() template.HTML {
				return csrf.TemplateField(r)
			},
			"currentUser": func() *models.User {
				return context.User(r.Context())
			},
		},
	)
	w.Header().Set("content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil{
		log.Printf("executing  template : %v", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return 
	}
	io.Copy(w, &buf)
}