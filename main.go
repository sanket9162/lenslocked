package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sanket9162/lenslocked/controllers"
	"github.com/sanket9162/lenslocked/models"
	"github.com/sanket9162/lenslocked/templates"
	"github.com/sanket9162/lenslocked/views"
)





func notFound(w http.ResponseWriter, r *http.Request){
	http.Error(w, "Page not Found", http.StatusNotFound)
}



func main(){

	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	} 
	defer db.Close()
	userService := models.USerService{
		DB: db,
	}

	r := chi.NewRouter()
	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))))
	r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))))

	userC :=  controllers.Users{
		USerService: &userService,
	}
	userC.Templates.New = views.Must(views.ParseFS(templates.FS,"signup.gohtml", "tailwind.gohtml" ))
	r.Get("/signup", userC.New)
	r.Post("/users", userC.Create)


	r.NotFound(notFound)


	fmt.Println("staring the server on :3000")
	http.ListenAndServe(":3000",r)

}