package user

import (
	"errors"
	"time"
)

var (
	ErrInvalid      = errors.New("invalid input")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("conflict")
	ErrNotFound     = errors.New("user not found")
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Credentials struct {
	Email    string
	Password string
}

type TokenPair struct {
	Access  string
	Refresh string
}
