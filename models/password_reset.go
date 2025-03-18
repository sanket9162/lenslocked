package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/sanket9162/lenslocked/rand"
)

const (
	DefaultResetDuration = 1 * time.Hour
)

type PasswordReset struct{
	ID int
	UserID int
	Token string
	TokenHash string
	ExpiresAt time.Time
}

type PasswordResetService struct{
	DB *sql.DB
	BytesPerToken int 
	Duration time.Duration
}

func (service *PasswordResetService) Create(email string) (*PasswordReset, error) {
	email = strings.ToLower(email)
	var userId int 
	row := service.DB.QueryRow(`
	SELECT id  FROM users WHERE email = $1;`, email)
	err := row.Scan(&userId)
	if err != nil {
		return nil, fmt.Errorf("creaete %w", err)
	}

	// Build the passwordReset
	bytesPerToken := service.BytesPerToken
	if bytesPerToken < MinBytesPerToken{
		bytesPerToken = MinBytesPerToken
	}
	token, err := rand.String(bytesPerToken) 
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	duration := service.Duration
	if duration == 0 {
		duration = DefaultResetDuration
	}
	pwReset := PasswordReset{
		UserID: userId,
		Token: token,
		TokenHash: service.hash(token),
		ExpiresAt: time.Now().Add(duration),
	}

	row = service.DB.QueryRow(`
	INSERT INTO password_resets (user_id, token_hash, expires_at)
	VALUES ($1, $2, $3) ON CONFLICT (user_id) DO
	UPDATE
	SET token_hash =$2, expires_at = $3
	RETURNING id;`, pwReset.UserID, pwReset.TokenHash, pwReset.ExpiresAt)
	err = row.Scan(&pwReset.ID)

	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}

	return &pwReset, nil
}

func (service *PasswordResetService) Consume(token string) (*User, error){
	return nil, fmt.Errorf("")
}

func (server *PasswordResetService) hash(token string) string{
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
 }