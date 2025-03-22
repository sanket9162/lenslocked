package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sanket9162/lenslocked/context"
	"github.com/sanket9162/lenslocked/models"
)

type Galleries struct{
	Templates struct{
		New Template
		Edit Template
	}
	GalleryService *models.GalleryService
}

func (g *Galleries) New(w http.ResponseWriter, r *http.Request){
	var data struct{
		Title string
	}
	data.Title = r.FormValue("title")
	g.Templates.New.Execute(w, r, data)
}

func (g *Galleries) Create(w http.ResponseWriter, r *http.Request){
	var data struct {
		UserID int 
		Title string
	}
	data.UserID = context.User(r.Context()).ID
	data.Title = r.FormValue("title")

	gallery, err := g.GalleryService.Create(data.Title, data.UserID)
	if err != nil{
		fmt.Print(err)
		g.Templates.New.Execute(w, r, data, err)
		return
		
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) Edit (w http.ResponseWriter, r *http.Request){
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return
	}
	gallery, err := g.GalleryService.ByID(id)
	if err != nil{
		if errors.Is(err, models.ErrNotFound){
		http.Error(w, "Gallery not foung", http.StatusNotFound)
		return
	}

	http.Error(w, "something went wrong", http.StatusInternalServerError)
	return
}
user := context.User(r.Context())
if gallery.UserID != user.ID {
	http.Error(w, "You are not authorized to edit this gallery", http.StatusForbidden)
	return
}
var data struct {
	ID int 
	Title string
}

data.ID = gallery.ID
data.Title = gallery.Title
g.Templates.Edit.Execute(w, r, data)

}