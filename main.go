package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/sanket9162/lenslocked/views"
)

func executeTemplate(w http.ResponseWriter, filepath string){

	t , err := views.Parse(filepath)
	if err != nil{
		log.Printf("parsing template : %v", err)
		http.Error(w, "There was an error parsing the template.", http.StatusInternalServerError)
		return
	}

	t.Execute(w, nil)
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