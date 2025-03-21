package views

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path"

	"github.com/gorilla/csrf"
	"github.com/sanket9162/lenslocked/context"
	"github.com/sanket9162/lenslocked/models"
)

type public interface {
	Public() string
}



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
	tpl := template.New(path.Base(pattern[0]))
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfFeild": func() (template.HTML, error) {
				return "", fmt.Errorf("csrfField not implemented")
			},
			"currentUser": func() (template.HTML, error) {
				return "", fmt.Errorf("csrfField not implemented")
			},
			"errors": func() []string{
				return nil
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



func (t Template) Execute (w http.ResponseWriter, r *http.Request, data interface{}, errs ...error){
	tpl, err := t.htmlTpl.Clone()
	if err != nil {
		log.Printf("Cloning template: %v", err)
		http.Error(w, "There was an error rendering the page.", http.StatusInternalServerError)
		return
	}
	errMsgs := errMessages(errs...)
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfFeild": func() template.HTML {
				return csrf.TemplateField(r)
			},
			"currentUser": func() *models.User {
				return context.User(r.Context())
			},
			"errors": func() []string{
				return errMsgs
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

func errMessages(errs ...error)[]string{
	var msg []string
	for _, err := range errs{
		var pubErr public
		if errors.As(err, &pubErr){
			msg = append(msg, pubErr.Public())
		} else {
			msg = append(msg, "something went wrong.")
		}
	}
	return msg
}