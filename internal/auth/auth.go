package auth

import "errors"

var ErrUnauthorized = errors.New("unauthorized")

type TokenPair struct {
	Access  string
	Refresh string
}
