package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ishowdarkside/go-movies-app/internal/validator"
	"github.com/lib/pq"
)

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *Movie) error {

	query := `INSERT INTO movies (title, year, runtime, genres) VALUES ($1, $2, $3, $4) RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	res := m.DB.QueryRowContext(ctx, query, movie.Title, movie.Year, movie.Runtime, pq.StringArray(movie.Genres))

	return res.Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}
	moviePlaceholder := Movie{}

	query := `SELECT  id, title, genres, runtime, year, created_at, version  FROM movies WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(&moviePlaceholder.ID, &moviePlaceholder.Title, pq.Array(&moviePlaceholder.Genres), &moviePlaceholder.Runtime, &moviePlaceholder.Year, &moviePlaceholder.CreatedAt, &moviePlaceholder.Version)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err

	}
	return &moviePlaceholder, err
}

func (m MovieModel) Update(movie *Movie) error {

	if movie.ID < 1 {
		return ErrRecordNotFound
	}

	query := `UPDATE movies SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1 
	WHERE id = $5 AND version = $6
	RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID, movie.Version).Scan(&movie.Version)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return ErrEditConflict
	}
	return err
}

func (m MovieModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM movies where id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	res, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitzero"`
	Runtime   Runtime   `json:"runtime,omitzero"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, m *Movie) bool {

	v.Check(m.Title != "", "title", "must be provided")
	v.Check(len(m.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(m.Year != 0, "year", "must be provided")
	v.Check(m.Year >= 1888 && m.Year <= int32(time.Now().Year()), "year", "must be greater than 1888 and not be in the future")

	v.Check(m.Runtime != 0, "runtime", "must be provided")
	v.Check(m.Runtime > 0, "runtime", "must be positive number")

	v.Check(m.Genres != nil, "genres", "must be provided")
	v.Check(len(m.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(m.Genres) <= 5, "genres", "must not contain more that 5 genres")
	v.Check(validator.Unique(m.Genres), "genres", "must not contain duplicate values")

	return v.Valid()

}
