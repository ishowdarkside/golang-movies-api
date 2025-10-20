package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"time"

	"github.com/ishowdarkside/go-movies-app/internal/validator"
)

const (
	ScopeActivation = "activation"
)

type Token struct {
	PlainText string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

func generateToken(userID int64, ttl time.Duration, scope string) *Token {

	token := &Token{

		PlainText: rand.Text(),
		UserID:    userID,
		Expiry:    time.Now().Add(ttl),
		Scope:     scope,
	}

	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token

}

func ValidatePlainText(v *validator.Validator, tokenPlainText string) {

	v.Check(tokenPlainText != "", "token", "must be provided")
	v.Check(len(tokenPlainText) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	DB *sql.DB
}

func (m *TokenModel) Insert(token *Token) error {

	query := `
	INSERT INTO tokens (hash, user_id, expiry, scope) 
	VALUES ($1, $2, $3, $4)`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, token.Hash, token.UserID, token.Expiry, token.Scope)
	return err

}

func (m *TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {

	token := generateToken(userID, ttl, scope)
	err := m.Insert(token)

	return token, err

}

func (m *TokenModel) DeleteAllForUser(scope string, userID int64) error {

	query := `
	DELETE FROM tokens WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err

}
