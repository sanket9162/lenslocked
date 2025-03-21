package models

import "errors"

var (
	ErrNotFound = errors.New("models: resourse could be not found")
	ErrEmailTaken = errors.New("models: email address is already in use")
)
