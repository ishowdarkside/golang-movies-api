package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/ishowdarkside/go-movies-app/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) GetForToken(scope string, tokenPlainText string) (*User, error) {

	var user User

	query := `SELECT users.id, users.created_at, users.name, users.password_hash, users.activated, users.email, users.version FROM USERS
	INNER JOIN tokens ON users.id = tokens.user_id WHERE tokens.scope = $1 AND tokens.hash = $2 and tokens.expiry > $3`

	tokenHash := sha256.Sum256([]byte(tokenPlainText))

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, scope, tokenHash[:], time.Now()).Scan(&user.ID, &user.CreatedAt, &user.Name, &user.Password.hash, &user.Activated, &user.Email, &user.Version)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err

	}

	return &user, nil

}

func (m *UserModel) Insert(user *User) error {

	query := `INSERT INTO users (name, email, password_hash, activated) VALUES ($1, $2, $3, $4) RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, user.Name, user.Email, user.Password.hash, user.Activated).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {

		if strings.Contains(err.Error(), `violates unique constraint "users_email_key"`) {
			return ErrDuplicateEmail
		}

		return err
	}

	return nil
}

func (m *UserModel) GetByEmail(email string) (*User, error) {

	input := User{}

	query := `SELECT id, created_at, name, email, password_hash, activated, version FROM users WHERE email = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(&input.ID, &input.CreatedAt, &input.Name, &input.Email, &input.Password.hash, &input.Activated, &input.Version)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}

		return nil, err
	}

	return &input, nil
}

func (m *UserModel) Update(user *User) error {

	query := `UPDATE users SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1 
	WHERE id = $5 AND version = $6 RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, user.Name, user.Email, user.Password.hash, user.Activated, user.ID, user.Version).Scan(&user.Version)

	if err != nil {

		if strings.Contains(err.Error(), `violates unique constraint "users_email_key"`) {
			return ErrDuplicateEmail
		}

		return err
	}

	return nil
}

func ValidateEmail(v *validator.Validator, email string) {

	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")

}

func ValidatePasswordPlaintext(v *validator.Validator, plainTextPassword string) {

	v.Check(plainTextPassword != "", "password", "must be provided")
	v.Check(len(plainTextPassword) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(plainTextPassword) <= 72, "password", "must not be more than 72 bytes long")

}

func ValidateUser(v *validator.Validator, user *User) {

	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")
	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

var AnonymousUser = &User{}

func (u *User) IsAnonymous() bool {

	return u == AnonymousUser

}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plainTextPassword string) error {

	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plainTextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plainTextPassword string) (bool, error) {

	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainTextPassword))
	if err != nil {

		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {

			return false, nil
		}
		return false, err
	}

	return true, nil

}
