package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func executeTemplate(w http.ResponseWriter, filepath string){
	w.Header().Set("content-Type", "text/html; charset=utf-8")
	tpl, err := template.ParseFiles(filepath)
	if err != nil{
		log.Printf("parsing template : %v", err)
		http.Error(w, "There was an error parsing the template.", http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, nil)
	if err != nil{
		log.Printf("executing  template : %v", err)
		http.Error(w, "There was an error executing the template.", http.StatusInternalServerError)
		return 
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request){
	tplPath := filepath.Join("templates", "home.gohtml")
	executeTemplate(w, tplPath)
}

func contactHandler(w http.ResponseWriter, r *http.Request){

	tplPath := filepath.Join("templates", "contact.gohtml")
	executeTemplate(w, tplPath)
}

func faqHandler(w http.ResponseWriter, r *http.Request){
	tplPath := filepath.Join("templates", "faq.gohtml")
	executeTemplate(w, tplPath)

}

func notFound(w http.ResponseWriter, r *http.Request){
	http.Error(w, "Page not Found", http.StatusNotFound)
}

func main(){
	r := chi.NewRouter()
	r.Get("/", homeHandler)
	r.Get("/contact", contactHandler)
	r.Get("/faq", faqHandler)
	r.NotFound(notFound)


	fmt.Println("staring the server on :3000")
	http.ListenAndServe(":3000",r)

}