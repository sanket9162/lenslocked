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

	fmt.Fprint(w, "Email: ", r.FormValue("email"))
	fmt.Fprint(w, "Password: ", r.FormValue("password"))

}