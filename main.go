package main

import (
	"fmt"
	"net/http"
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


func main(){
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/contact", contactHandler)
	http.HandleFunc("/faq", faqHandler)

	fmt.Println("staring the server on :3000")
	http.ListenAndServe(":3000",nil)

}