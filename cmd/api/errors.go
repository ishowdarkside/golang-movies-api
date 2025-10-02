package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {

	var (
		method = r.Method
		uri    = r.RequestURI
	)

	app.logger.Error(err.Error(), slog.String("method", method), slog.String("uri", uri))
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {

	env := envelope{"error": message}
	err := app.writeJSON(w, status, env)

	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {

	app.logError(r, err)
	var (
		method = r.Method
		uri    = r.RequestURI
	)

	if app.config.env == "development" {
		message := envelope{"message": err.Error(), "method": method, "uri": uri}
		app.errorResponse(w, r, http.StatusInternalServerError, message)
	} else {
		message := "the server encountered a problem and could not process your request"
		app.errorResponse(w, r, http.StatusInternalServerError, message)
	}

}

func (app *application) notFoundError(w http.ResponseWriter, r *http.Request) {

	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)

}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {

	method := r.Method
	message := fmt.Sprintf("method %s is not allowed for this resource", method)

	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.errorResponse(w, r, http.StatusBadRequest, err.Error())

}
