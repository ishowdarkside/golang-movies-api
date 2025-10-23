package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ishowdarkside/go-movies-app/internal/data"
	"github.com/ishowdarkside/go-movies-app/internal/validator"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

func (app *application) rateMilit(next http.Handler) http.Handler {

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {

		for {

			time.Sleep(time.Minute)
			mu.Lock()

			for ip, client := range clients {

				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}

			}

			mu.Unlock()

		}

	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !app.config.limiter.enabled {

			next.ServeHTTP(w, r)
			return
		}
		ip := realip.FromRequest(r)
		mu.Lock()

		if clients[ip] == nil {
			clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst), lastSeen: time.Now()}
		}

		if !clients[ip].limiter.Allow() {

			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return

		}

		mu.Unlock()

		next.ServeHTTP(w, r)

	})

}

func (app *application) recoverPanic(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {

			if err := recover(); err != nil {

				w.Header().Set("Connection", "close")
				app.serverError(w, r, fmt.Errorf("%s", err))
				return
			}

		}()

		next.ServeHTTP(w, r)

	})

}

func (app *application) authenticate(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {

			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return

		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {

			app.invalidAuthenicationTokenResponse(w, r)
			return

		}

		token := headerParts[1]
		v := validator.New()

		if data.ValidateTokenPlainText(v, token); !v.Valid() {

			app.invalidAuthenicationTokenResponse(w, r)
			return

		}

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {

			if errors.Is(err, data.ErrRecordNotFound) {

				app.invalidAuthenicationTokenResponse(w, r)
				return
			}

			app.serverError(w, r, err)
			return

		}

		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)

	})
}

func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		user := app.contextGetUser(r)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)

	})

}
