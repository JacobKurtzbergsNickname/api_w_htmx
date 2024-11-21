package main

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// LoggingMiddleware logs incoming requests along with the assigned emoji.
func LoggingMiddleware(tl *TableLogger) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Retrieve the emoji from the context
			emoji, ok := r.Context().Value(emojiKey).(string)
			if !ok || emoji == "" {
				emoji = "🤷‍♂️" // Default shrugging man emoji if not found
			}

			// Skip any requests to the static files in logging
			if strings.Contains(r.URL.Path, "/static/") {
				next.ServeHTTP(w, r)
				return
			}

			// Log the request details along with the emoji
			tl.LogRequest(emoji, r.Method, r.URL.Path, r.RemoteAddr)

			// Proceed to the next middleware/handler
			next.ServeHTTP(w, r)
		}
	}
}

// EmojiMiddleware assigns an emoji based on the time of day.
// There's a 10% chance it assigns no time-based emoji, triggering the fallback.
func EmojiMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Seed the random number generator
		randomNumber := rand.New(rand.NewSource(time.Now().UnixNano()))

		// 10% chance to skip time-based emoji assignment
		skip := randomNumber.Intn(100) < 10 // 10% probability

		var emoji string

		if !skip {
			// Determine the current time of day
			now := time.Now().In(time.UTC) // Adjust to your timezone if necessary
			hour := now.Hour()

			switch {
			case hour >= 6 && hour < 12:
				emoji = "☀️" // Morning
			case hour >= 12 && hour < 18:
				emoji = "🌞" // Afternoon
			case hour >= 18 && hour < 21:
				emoji = "🌇" // Evening
			case hour >= 21 || hour < 6:
				emoji = "🌙" // Night
			default:
				// This default should never be hit, but just in case
				emoji = "🤷‍♂️" // Shrugging man emoji
			}
		}

		if emoji == "" {
			// Assign the shrugging man emoji as a fallback
			emoji = "🤷‍♂️"
		}

		// Store the emoji in the request context
		ctx := context.WithValue(r.Context(), emojiKey, emoji)
		r = r.WithContext(ctx)

		// Proceed to the next middleware/handler
		next.ServeHTTP(w, r)
	})
}

func RequireAuthMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "Bearer token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}

type Middleware func(http.Handler) http.HandlerFunc

func ChainMiddleware(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}

		return next.ServeHTTP
	}
}
