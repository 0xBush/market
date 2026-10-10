package order

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("order not found")

type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Msg
}

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
