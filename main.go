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
	//TODO : Read all values from an env variable
	cfg.PSQL = models.DefaultPostgresConfig()

	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMPT_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil{
		return cfg, err
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")


	cfg.CSRF.Key = "gFvi45R4fy5xNBlnEeZtQbfAVCYEIAUX"
	cfg.CSRF.Secure = false
	cfg.Server.Address = ":3000"
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
	emailService := models.NewEmailService(cfg.SMTP)

	//Setup middleware
	umw := controllers.Usermiddleware{
		SessionService: sessionService,
	}

	csrfMw := csrf.Protect(
		[]byte(cfg.CSRF.Key),
		csrf.Secure(cfg.CSRF.Secure),
	)

	//Setup controllers
	userC :=  controllers.Users{
		UserService: userService,
		SessionService: sessionService,
		PasswordResetService: pwResetService,
		EmailService: emailService,
	}
	userC.Templates.New = views.Must(views.ParseFS(templates.FS,"signup.gohtml", "tailwind.gohtml" ))
	userC.Templates.SignIn  = views.Must(views.ParseFS(templates.FS,"signin.gohtml", "tailwind.gohtml" ))
	userC.Templates.ForgotPassword  = views.Must(views.ParseFS(templates.FS,"forgot-pw.gohtml", "tailwind.gohtml" ))
	userC.Templates.CheckYourEmail  = views.Must(views.ParseFS(templates.FS,"check-your-email.gohtml", "tailwind.gohtml" ))


	

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
	fmt.Printf("staring the server on %s...\n", cfg.Server.Address)
	err = http.ListenAndServe(cfg.Server.Address, r)
	if err != nil {
		panic(err)
	}
}