package main

import (
	"context"
	"net/http"

	"github.com/ishowdarkside/go-movies-app/internal/data"
)

type contextKey string

const userContextKy = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {

	ctx := context.WithValue(r.Context(), userContextKy, user)
	return r.WithContext(ctx)

}

func (app *application) contextGetUser(r *http.Request) *data.User {

	user, ok := r.Context().Value(userContextKy).(*data.User)

	if !ok {
		panic("missing user valuei n request context")
	}

	return user

}
