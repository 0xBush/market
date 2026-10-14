package httputil

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrInvalidJSON  = errors.New("invalid JSON body")
	ErrBodyTooLarge = errors.New("request body too large")
)

type ErrorBody struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, ErrorBody{Error: msg})
}

func WriteReadError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrBodyTooLarge) {
		WriteError(w, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	WriteError(w, http.StatusBadRequest, ErrInvalidJSON.Error())
}

func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	const limit = 1 << 20

	r.Body = http.MaxBytesReader(w, r.Body, limit)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return ErrBodyTooLarge
		}
		return ErrInvalidJSON
	}

	return nil
}
