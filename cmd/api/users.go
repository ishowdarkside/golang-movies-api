package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/ishowdarkside/go-movies-app/internal/data"
	"github.com/ishowdarkside/go-movies-app/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {

		app.serverError(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateUser(v, user)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {

		if errors.Is(err, data.ErrDuplicateEmail) {

			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
		app.serverError(w, r, err)
		return
	}

	token, err := app.models.Tokens.New(user.ID, time.Hour*24*3, data.ScopeActivation)
	if err != nil {

		app.serverError(w, r, err)
		return
	}

	app.background(func() {

		data := map[string]any{
			"userID":          user.ID,
			"activationToken": token.PlainText,
		}
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.Error(err.Error())
		}

	})

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverError(w, r, err)
	}

}
