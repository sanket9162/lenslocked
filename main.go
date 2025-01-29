package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)


func homeHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1> hello from go</h1>")
}

func contactHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w,"<h1> hello from Contact page</h1>" )
}

func faqHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w , `<h1>FAQ Page</h1>
	
`)
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