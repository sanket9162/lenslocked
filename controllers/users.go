package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/sanket9162/lenslocked/context"
	"github.com/sanket9162/lenslocked/errors"
	"github.com/sanket9162/lenslocked/models"
)

type Users struct{
	Templates struct {
		New Template
		SignIn Template
		ForgotPassword Template
		CheckYourEmail Template
		ResetPassword Template
	}
	UserService *models.UserService
	SessionService *models.SessionService
	PasswordResetService *models.PasswordResetService
	EmailService *models.EmailService
}

func (u Users) New(w http.ResponseWriter, r *http.Request){
	var data struct{
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.New.Execute(w, r, data)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request){
	var data struct{
		Email string
		Password string
	}
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")
	user, err := u.UserService.Create(data.Email, data.Password)
	if err != nil {
		if errors.Is(err, models.ErrEmailTaken){
			err = errors.Public(err, "This email address is already associated with an account.")
		}
		u.Templates.New.Execute(w, r, data, err)
		return
	}
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/galleries", http.StatusFound)
}


func (u Users) SignIn(w http.ResponseWriter, r *http.Request){
	var data struct{
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.SignIn.Execute(w, r, data)
}

func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request){
	var data struct{
		Email string
		Password string 
	}
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")
	user, err := u.UserService.Authenticate(data.Email, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "something went wrong ", http.StatusInternalServerError)
		return

	}
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "something went wrong ", http.StatusInternalServerError)
		return
	}
	setCookie(w, CookieSession, session.Token)

	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request){
	user := context.User(r.Context())
	fmt.Fprintf(w, "current user: %s\n", user.Email)

}

func (u Users) ProcessSignOut(w http.ResponseWriter, r *http.Request ){
	token, err:= readCookie(r, CookieSession)
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound) 
		return
	}
	err = u.SessionService.Delete(token)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	deleteCookie(w, CookieSession)
	http.Redirect(w, r, "/signin", http.StatusFound)
}

func (u Users) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.ForgotPassword.Execute(w, r, data)
}

func (u Users) ProcessForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	pwReset, err := u.PasswordResetService.Create(data.Email)
	if err != nil {
		// TODO: Handle other cases in the future. For instance, if a user does not exist with that email address.
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	vals := url.Values{
		"token": {pwReset.Token},
	}
	resetURL := "https://www.lenslocked.com/reset-pw?" + vals.Encode()
	err = u.EmailService.ForgotPassword(data.Email, resetURL)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	u.Templates.CheckYourEmail.Execute(w, r, data)
}


func (u Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string
	}
	data.Token = r.FormValue("token")
	u.Templates.ResetPassword.Execute(w, r, data)
}

func (u Users) ProcessResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token    string
		Password string
	}
	data.Token = r.FormValue("token")
	data.Password = r.FormValue("password")

	user, err := u.PasswordResetService.Consume(data.Token)
	if err != nil {
		fmt.Println(err)
		// TODO: Distinguish between types of errors.
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	err = u.UserService.UpdatePassword(user.ID, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// Sign the user in now that their password has been reset.
	// Any errors from this point onwards should redirect the user to the sign in
	// page.
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

type Usermiddleware struct{
	SessionService *models.SessionService
}

func (umw Usermiddleware) SetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
	token, err:= readCookie(r, CookieSession)
	if err != nil {
		next.ServeHTTP(w, r)
		return
	}
	user, err := umw.SessionService.User(token)
	if err != nil {
		next.ServeHTTP(w, r)
		return
	}
	ctx := r.Context()
	ctx = context.WithUser(ctx, user)
	r = r.WithContext(ctx)
	next.ServeHTTP(w, r)
	})

}

func (umw Usermiddleware) RequestUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}
		next.ServeHTTP(w,r)
	})
}