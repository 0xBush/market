package order

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("order not found")
	ErrInvalid  = errors.New("invalid input")
)

type Order struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Quantity  int32     `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInput struct {
	Name     string
	Quantity int32
}
