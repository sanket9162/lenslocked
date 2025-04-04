package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
	"github.com/sanket9162/lenslocked/controllers"
	"github.com/sanket9162/lenslocked/migrations"
	"github.com/sanket9162/lenslocked/models"
	"github.com/sanket9162/lenslocked/templates"
	"github.com/sanket9162/lenslocked/views"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMPTConfig
	CSRF struct{
		Key string
		Secure bool
	}
	Server struct{
		Address string
	}
}

func loadEnvConfig() (config, error){
	var cfg config
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}

	cfg.PSQL = models.PostgresConfig{
		Host:     os.Getenv("PSQL_HOST"),
		Port:     os.Getenv("PSQL_PORT"),
		User:     os.Getenv("PSQL_USER"),
		Password: os.Getenv("PSQL_PASSWORD"),
		Database: os.Getenv("PSQL_DATABASE"),
		SSLMode:  os.Getenv("PSQL_SSLMODE"),
	}
	if cfg.PSQL.Host == "" && cfg.PSQL.Port == "" {
		return cfg, fmt.Errorf("No PSQL config provided.")
	}

	cfg.SMTP.Host = os.Getenv("SMPT_HOST")
	portStr := os.Getenv("SMPT_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil{
		return cfg, err
	}
	cfg.SMTP.Username = os.Getenv("SMPT_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMPT_PASSWORD")


	cfg.CSRF.Key = os.Getenv("CSRF_KEY")
	cfg.CSRF.Secure = os.Getenv("CSRF_SECURE") == "true"
	cfg.Server.Address = os.Getenv("SERVER_ADDRESS")
	return cfg, nil
}

func main(){
	cfg , err :=  loadEnvConfig()
	if err != nil{
		panic(err)
	}

	// Setup the Database
	db, err := models.Open(cfg.PSQL)
	if err != nil {
		panic(err)
	} 
	defer db.Close()

	err = models.MigrateFS(db, migrations.FS,".")
	if err != nil{
		panic(err)
	}

	//Setup services

	userService := &models.UserService{
		DB: db,
	}
	sessionService := &models.SessionService{
		DB: db,
	}
	pwResetService := &models.PasswordResetService{
		DB: db,
	}
	galleryService := &models.GalleryService{
		DB: db,
	}
	emailService := models.NewEmailService(cfg.SMTP)

	//Setup middleware
	umw := controllers.Usermiddleware{
		SessionService: sessionService,
	}

	csrfMw := csrf.Protect(
		[]byte(cfg.CSRF.Key),
		csrf.Secure(cfg.CSRF.Secure),
		csrf.Path("/"),
	)

	//Setup controllers
	userC :=  controllers.Users{
		UserService: userService,
		SessionService: sessionService,
		PasswordResetService: pwResetService,
		EmailService: emailService,
	}
	galleriesC := controllers.Galleries{
		GalleryService: galleryService,
	}
	userC.Templates.New = views.Must(views.ParseFS(templates.FS,"signup.gohtml", "tailwind.gohtml" ))
	userC.Templates.SignIn  = views.Must(views.ParseFS(templates.FS,"signin.gohtml", "tailwind.gohtml" ))
	userC.Templates.ForgotPassword  = views.Must(views.ParseFS(templates.FS,"forgot-pw.gohtml", "tailwind.gohtml" ))
	userC.Templates.CheckYourEmail  = views.Must(views.ParseFS(templates.FS,"check-your-email.gohtml", "tailwind.gohtml" ))
	userC.Templates.ResetPassword  = views.Must(views.ParseFS(templates.FS,"reset-pw.gohtml", "tailwind.gohtml" ))
	galleriesC.Templates.New = views.Must(views.ParseFS(templates.FS,"galleries/new.gohtml", "tailwind.gohtml",))
	galleriesC.Templates.Edit = views.Must(views.ParseFS(templates.FS,"galleries/edit.gohtml", "tailwind.gohtml",))
	galleriesC.Templates.Index = views.Must(views.ParseFS(templates.FS,"galleries/index.gohtml", "tailwind.gohtml",))
	galleriesC.Templates.Show = views.Must(views.ParseFS(templates.FS,"galleries/show.gohtml", "tailwind.gohtml",))


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
	r.Post("/forgot-pw", userC.ProcessForgotPassword)
	r.Get("/reset-pw", userC.ResetPassword)
	r.Post("/reset-pw", userC.ProcessResetPassword)
	r.Route("/users/me", func(r chi.Router){
		r.Use(umw.RequestUser)
		r.Get("/", userC.CurrentUser)
	})
	r.Route("/galleries", func(r chi.Router) {
		r.Get("/{id}", galleriesC.Show)
		r.Get("/{id}/images/{filename}", galleriesC.Images)
		r.Group(func(r chi.Router) {
			r.Use(umw.RequestUser)
			r.Get("/", galleriesC.Index)
			r.Get("/new", galleriesC.New)
			r.Post("/", galleriesC.Create)
			r.Get("/{id}/edit", galleriesC.Edit)
			r.Post("/{id}", galleriesC.Update)
			r.Post("/{id}/delete", galleriesC.Delete)
			r.Post("/{id}/images", galleriesC.UploadImages)
			r.Post("/{id}/images/{filename}/delete", galleriesC.DeleteImages)
		})
	})
	
	r.NotFound(func(w http.ResponseWriter, r *http.Request){
		http.Error(w, "Page not Found", http.StatusNotFound)
	})

	//start the server
	fmt.Printf("staring the server on %s...\n", cfg.Server.Address)
	err = http.ListenAndServe(cfg.Server.Address, r)
	if err != nil {
		panic(err)
	}
}