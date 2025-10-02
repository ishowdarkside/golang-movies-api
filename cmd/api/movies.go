package main

import (
	"net/http"
	"time"

	"github.com/ishowdarkside/go-movies-app/internal/data"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year,omitempty"`
		Runtime int32    `json:"runtime,omitempty"`
		Genres  []string `json:"genres,omitempty"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {

		app.badRequestResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, envelope{"movie": input})

}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)

	if err != nil || id < 1 {

		app.notFoundError(w, r)
		return
	}

	movieInstance := data.Movie{ID: id, Title: "Spider-Man 2004", CreatedAt: time.Now(), Runtime: 102, Genres: []string{"fantasy", "science fiction"}, Version: 1}
	err = app.writeJSON(w, 200, envelope{"movie": movieInstance})

	if err != nil {
		app.serverError(w, r, err)
	}

}
