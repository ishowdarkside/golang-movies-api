package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ishowdarkside/go-movies-app/internal/data"
	"github.com/ishowdarkside/go-movies-app/internal/validator"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year,omitempty"`
		Runtime data.Runtime `json:"runtime,omitempty"`
		Genres  []string     `json:"genres,omitempty"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {

		app.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Genres:  input.Genres,
		Year:    input.Year,
		Runtime: input.Runtime,
	}

	v := validator.New()
	isMovieValid := data.ValidateMovie(v, movie)

	if !isMovieValid {

		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {

		app.serverError(w, r, err)
	}

}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)

	if err != nil || id < 1 {

		app.notFoundError(w, r)
		return
	}

	movieInstance, err := app.models.Movies.Get(id)

	if err != nil {

		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundError(w, r)
			return
		}

		app.serverError(w, r, err)
		return

	}

	err = app.writeJSON(w, 200, envelope{"movie": movieInstance}, nil)

	if err != nil {
		app.serverError(w, r, err)
	}

}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundError(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {

		if errors.Is(err, data.ErrRecordNotFound) {

			app.notFoundError(w, r)
			return
		}

		app.serverError(w, r, err)
		return
	}

	if r.Header.Get("X-Expected-Version") != "" {

		if strconv.Itoa(int(movie.Version)) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}

	}

	var input struct {
		Title   *string       `json: "title"`
		Year    *int32        `json: "year"`
		Runtime *data.Runtime `json: "runtime"`
		Genres  []string      `json: "genres"`
	}

	err = app.readJSON(w, r, &input)

	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}

	v := validator.New()
	if !data.ValidateMovie(v, movie) {

		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Update(movie)
	if err != nil {

		if errors.Is(err, data.ErrEditConflict) {
			app.editConflictResponse(w, r)
			return
		}

		app.serverError(w, r, err)
		return
	}
	app.writeJSON(w, 200, envelope{"movie": movie}, nil)

}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {

		app.notFoundError(w, r)
		return
	}

	errorRemovingRecord := app.models.Movies.Delete(id)
	if errorRemovingRecord != nil {

		if errors.Is(errorRemovingRecord, data.ErrRecordNotFound) {
			app.notFoundError(w, r)
			return
		}

		app.serverError(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusNoContent, envelope{}, nil)
	if err != nil {
		app.serverError(w, r, err)
	}

}
