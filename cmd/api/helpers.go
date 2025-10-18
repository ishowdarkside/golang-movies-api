package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ishowdarkside/go-movies-app/internal/validator"
	"github.com/julienschmidt/httprouter"
)

type envelope map[string]any

func (app *application) readIDParam(r *http.Request) (int64, error) {

	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)

	if err != nil || id < 1 {

		return 0, errors.New("invalid id parameter")
	}

	return id, nil

}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {

	js, err := json.Marshal(data)

	if err != nil {
		return err
	}

	for headerItem := range headers {
		fmt.Println(headers[headerItem])
		w.Header()[headerItem] = headers[headerItem]
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil

}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {

	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(dst)

	if err != nil {

		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {

		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be large than %d bytes", maxBytesError.Limit)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		default:
			return err

		}

	}

	err = decoder.Decode(r.Body)
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}
	return nil

}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {

	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	return s

}

func (app *application) readCSV(qs url.Values, key string, defaultValues []string) []string {

	val := qs.Get(key)
	if val == "" {
		return defaultValues
	}

	return strings.Split(val, ",")

}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {

	val := qs.Get(key)
	if val == "" {
		return defaultValue
	}

	numVal, err := strconv.ParseInt(val, 10, 64)

	if err != nil {
		v.AddError(key, "must be an integer")
		return defaultValue
	}

	return int(numVal)

}

func (app *application) background(fn func()) {

	go func() {

		defer func() {

			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}

		}()

		fn()

	}()
}
