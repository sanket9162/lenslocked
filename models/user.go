package models

import "database/sql"

type User struct{
	ID           int
	Email	     string
	PasswordHash string
}

type USerService struct{
	DB *sql.DB
}