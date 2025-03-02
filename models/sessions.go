package models

import (
	"database/sql"
	"fmt"

	"github.com/sanket9162/lenslocked/rand"
)
 type Session struct{
	ID int 
	UserID int 
	Token string
	TokenHash string
 }

 type SessionService struct{
	DB *sql.DB
 }

 func (ss *SessionService) Create(userID int) (*Session, error) {
	token, err := rand.SessionToken()
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	session := Session{
		UserID: userID,
		Token: token,
	}
	return &session,nil
 } 

 func (ss *SessionService) User(token string) (*User, error){
	return nil, nil
 }