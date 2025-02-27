package views

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
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
			"csrfFeild": func() template.HTML {
				return `<input type="hidden"/>`
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

func (t Template) Execute (w http.ResponseWriter, data interface{}){
	w.Header().Set("content-Type", "text/html; charset=utf-8")

	err := t.htmlTpl.Execute(w, data)
	if err != nil{
		log.Printf("executing  template : %v", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return 
	}
}