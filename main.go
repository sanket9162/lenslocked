package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/sanket9162/lenslocked/controllers"
	"github.com/sanket9162/lenslocked/migrations"
	"github.com/sanket9162/lenslocked/models"
	"github.com/sanket9162/lenslocked/templates"
	"github.com/sanket9162/lenslocked/views"
)


func main(){
	// Setup the Database
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	} 
	defer db.Close()

	err = models.MigrateFS(db, migrations.FS,".")
	if err != nil{
		panic(err)
	}

	//Setup services

	userService := models.UserService{
		DB: db,
	}
	sessionService := models.SessionService{
		DB: db,
	}

	//Setup middleware
	umw := controllers.Usermiddleware{
		SessionService: &sessionService,
	}

	csrfKey := "gFvi45R4fy5xNBlnEeZtQbfAVCYEIAUX"
	csrfMw := csrf.Protect(
		[]byte(csrfKey),
		// TODO: Fix this before deploying
		csrf.Secure(false),
	)

	//Setup controllers
	userC :=  controllers.Users{
		UserService: &userService,
		SessionService: &sessionService,
	}
	userC.Templates.New = views.Must(views.ParseFS(templates.FS,"signup.gohtml", "tailwind.gohtml" ))
	userC.Templates.SignIn  = views.Must(views.ParseFS(templates.FS,"signin.gohtml", "tailwind.gohtml" ))
	userC.Templates.ForgotPassword  = views.Must(views.ParseFS(templates.FS,"forgot-pw.gohtml", "tailwind.gohtml" ))

	

	//Setup our router and routes
	r := chi.NewRouter()
	r.Use(csrfMw)
	r.Use(umw.SetUser)
	r.Get("/", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))))
	r.Get("/faq", controllers.FAQ(views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))))
	r.Get("/signup", userC.New)
	r.Post("/users", userC.Create)
	r.Get("/signin", userC.SignIn)
	r.Post("/signin", userC.ProcessSignIn)
	r.Post("/signout", userC.ProcessSignOut)
	r.Get("/forgot-pw", userC.ForgotPassword)
	r.Post("/forgot-pw", userC.ProcessForgetPassword)
	r.Route("/users/me", func(r chi.Router){
		r.Use(umw.RequestUser)
		r.Get("/", userC.CurrentUser)
	})
	// r.Get("/users/me", userC.CurrentUser)
	r.NotFound(func(w http.ResponseWriter, r *http.Request){
		http.Error(w, "Page not Found", http.StatusNotFound)
	})

	//start the server
	fmt.Println("staring the server on :3000")
	http.ListenAndServe(":3000", r)

}