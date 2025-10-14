package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

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
