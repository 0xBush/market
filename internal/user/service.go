package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost    = 12
	minPassword   = 8
	maxPassword   = 72
	maxEmailRunes = 254
)

type repo interface {
	Create(ctx context.Context, email, passwordHash string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, string, error)
	GetByID(ctx context.Context, id int64) (User, error)
}

type Service struct {
	repo repo
}

func NewService(r repo) *Service {
	return &Service{repo: r}
}

func (s *Service) Register(ctx context.Context, in Credentials) (User, error) {
	email, password, err := validateCredentials(in)
	if err != nil {
		return User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return User{}, err
	}

	return s.repo.Create(ctx, email, string(hash))
}

func (s *Service) Authenticate(ctx context.Context, in Credentials) (User, error) {
	email, password, err := validateCredentials(in)
	if err != nil {
		return User{}, err
	}

	u, passwordHash, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return User{}, ErrNotFound
		}
		return User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return User{}, ErrNotFound
	}

	return u, nil
}

func (s *Service) GetByID(ctx context.Context, userID int64) (User, error) {
	return s.repo.GetByID(ctx, userID)
}

func validateCredentials(in Credentials) (string, string, error) {
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return "", "", err
	}
	if n := utf8.RuneCountInString(in.Password); n < minPassword || n > maxPassword {
		return "", "", fmt.Errorf("%w: password must be %d-%d characters", ErrInvalid, minPassword, maxPassword)
	}
	return email, in.Password, nil
}

func normalizeEmail(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	n := utf8.RuneCountInString(s)
	if n < 3 || n > maxEmailRunes {
		return "", fmt.Errorf("%w: invalid email", ErrInvalid)
	}
	at := strings.IndexByte(s, '@')
	if at < 1 || at != strings.LastIndexByte(s, '@') || at == len(s)-1 {
		return "", fmt.Errorf("%w: invalid email", ErrInvalid)
	}
	return s, nil
}

// dummyHash is a valid bcrypt hash used when the email is unknown so login
// timing is closer to the password-mismatch path.
var dummyHash = []byte("$2a$12$R9h/cIPz0gi.URNNX3kh2OPST9/PgBkqquzi.Ss7KIUgO2t0jWMUW")
