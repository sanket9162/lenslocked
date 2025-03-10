package controllers

import (
	"fmt"
	"net/http"

	"github.com/sanket9162/lenslocked/context"
	"github.com/sanket9162/lenslocked/models"
)

type Users struct{
	Templates struct {
		New Template
		SignIn Template
	}
	UserService *models.UserService
	SessionService *models.SessionService
}

func (u Users) New(w http.ResponseWriter, r *http.Request){
	var data struct{
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.New.Execute(w, r, data)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request){
	email := r.FormValue("email")
	password:= r.FormValue("password")
	user, err := u.UserService.Create(email, password)
	if err != nil {
		fmt.Print(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
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

	http.Redirect(w, r, "/users/me", http.StatusFound)
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