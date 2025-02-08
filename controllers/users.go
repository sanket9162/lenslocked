package controllers

import (
	"fmt"
	"net/http"

	"github.com/sanket9162/lenslocked/models"
)

type Users struct{
	Templates struct {
		New Template
	}
	USerService *models.USerService
}

func (u Users) New(w http.ResponseWriter, r *http.Request){
	var data struct{
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.New.Execute(w, data)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request){

	email := r.FormValue("email")
	password:= r.FormValue("password")
	user, err := u.USerService.Create(email, password)
	if err != nil {
		fmt.Print(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
		fmt.Fprintf(w, "user create: %+v", user)
}