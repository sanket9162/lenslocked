package models

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Images struct{
	GalleryID int
	Path string
	Filename string
}

type Gallery struct {
	ID int 
	UserID int
	Title string
}

type GalleryService struct{
	DB *sql.DB
	ImagesDir string
}

func (service *GalleryService) Create(title string, userID int) (*Gallery, error) {
	gallery := Gallery{
		Title:  title,
		UserID: userID,
	}
	row := service.DB.QueryRow(`
		INSERT INTO galleries (title, user_id)
		VALUES ($1, $2) RETURNING id;`, gallery.Title, gallery.UserID)
	err := row.Scan(&gallery.ID)
	if err != nil {
		return nil, fmt.Errorf("create gallery: %w", err)
	}
	return &gallery, nil
}

func (service *GalleryService) ByID(id int) (*Gallery, error) {
	gallery := Gallery{
		ID: id,
	}
	row := service.DB.QueryRow(`
		SELECT title, user_id
		FROM galleries
		WHERE id = $1;`, gallery.ID)
	err := row.Scan(&gallery.Title, &gallery.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows){
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query gallery by id:%w", err)
	}
	return &gallery, nil
}

func (service *GalleryService) ByUserId(userId int) ([]Gallery, error){
	rows, err := service.DB.Query(`
	SELECT id, title
	FROM galleries
	WHERE user_id = $1;`, userId)
	if err != nil{
		return nil, fmt.Errorf("query galleries by user: %w", err)
	}
	var galleries [] Gallery 
	for rows.Next(){
		gallery := Gallery{
			UserID: userId,
		}
		err = rows.Scan(&gallery.ID, &gallery.Title)
		if err != nil {
			return nil, fmt.Errorf("query galleries by user: %w", err )
		}
		galleries = append(galleries, gallery)
	}
	if rows.Err() != nil{
		return nil, fmt.Errorf("query galleryes by user: %w", err)
	}
	return galleries, nil

}

func (service *GalleryService) Update(gallery *Gallery) error {
	_, err := service.DB.Exec(`
		UPDATE galleries
		SET title = $2
		WHERE id = $1;`, gallery.ID, gallery.Title)
	if err != nil {
		return fmt.Errorf("update gallery: %w", err)
	}
	return nil
}

func (service *GalleryService) Delete(id int) error{
	_, err := service.DB.Exec(`
	DELETE FROM galleries
	WHERE id = $1;`, id)
	if err != nil {
		return fmt.Errorf("delete gallery: %w", err)
	}
	err = os.RemoveAll(service.galleryDir(id))
	if err != nil {
		return fmt.Errorf("delete gallery images: %w",err)
	}
	return nil
}

func(service *GalleryService) Images(galleryID int) ([]Images, error){
	globPattern := filepath.Join(service.galleryDir(galleryID), "*")
	allFiles, err := filepath.Glob(globPattern)
	if err != nil{
		return nil, fmt.Errorf("retrieving gallery images: %w", err)
	}
	var images []Images
	for _, file := range allFiles{
		if hasExtension(file, service.extensions()){
		images = append(images, Images{
			GalleryID: galleryID,
			Path:file,
			Filename: filepath.Base(file),
		})
	}
	}
	return images, nil
}

func (service *GalleryService) Image(galleryID int, filename string)(Images, error){
	imagePath := filepath.Join(service.galleryDir(galleryID), filename)
	_, err := os.Stat(imagePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist){
			return Images{}, ErrNotFound
		}
		return Images{}, fmt.Errorf("querying for image: %w", err)
	}
	return Images{
		Filename :filename,
		GalleryID: galleryID,
		Path: imagePath,
	}, nil
}

func (service *GalleryService) CreateImage(galleryID int, filename string, contents io.ReadSeeker)(error){
	err := checkContentType(contents, service.imageContentType())
	if err != nil{
		return fmt.Errorf("createing image %v: %w", filename, err)
	}
	err = checkExtension(filename, service.extensions())
	if err != nil{
		return fmt.Errorf("createing image %v: %w", filename, err)
	}

	galleryDir := service.galleryDir(galleryID)
	err = os.MkdirAll(galleryDir, 0755)
	if err != nil {
		return fmt.Errorf("creating gallery %d image directory: %w", galleryID, err)
	}

	imagePath := filepath.Join(galleryDir, filename)
	dst, err := os.Create(imagePath)
	if err != nil{
		return fmt.Errorf("create image file: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, contents)
	if err != nil {
		return fmt.Errorf("copying contents to image %w", err)
	}
	return nil
}


func (service *GalleryService) DeleteImage(galleryID int, filename string)(error){
	image, err := service.Image(galleryID, filename)
	if err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}
	err = os.Remove(image.Path)
	if err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}
	return nil
}

func(service *GalleryService) extensions() []string {
	return []string{".png", ".jpg", ".jpeg", ".gif" }
}

func(service *GalleryService) imageContentType() []string {
	return []string{"image/png", "image/jpeg",  "image/gif" }
}

func(service *GalleryService) galleryDir(id int) string{
	imagesDIR :=service.ImagesDir
	if imagesDIR == "" {
		imagesDIR = "images"
	}
	return filepath.Join(imagesDIR, fmt.Sprintf("gallery-%d", id))
}

func hasExtension(file string, extensions []string) bool {
	for _, ext := range extensions{
		file = strings.ToLower(file)
		ext = strings.ToLower(ext)
		if filepath.Ext(file) == ext {
			return true
		}
	}
	return false
}