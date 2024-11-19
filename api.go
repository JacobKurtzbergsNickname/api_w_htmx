package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// Define a custom type for context keys to avoid collisions
type contextKey string

const emojiKey contextKey = "emoji"

type APIServer struct {
	addr              string
	routes            []Route
	middlewareChain   Middleware
	activeMiddlewares []string
}

type Route struct {
	path   string
	method func(w http.ResponseWriter, r *http.Request)
}

func NewAPIServer(addr string) *APIServer {
	return &APIServer{addr: addr}
}

func (s *APIServer) AddRoute(route Route) {
	s.routes = append(s.routes, route)
}

func (s *APIServer) AddRoutes(routes []Route) {
	s.routes = append(s.routes, routes...)
}

func (s *APIServer) AddMiddleware(middlewares []string) {
	s.activeMiddlewares = middlewares
}

func (s *APIServer) InitRouter() *http.ServeMux {
	router := http.NewServeMux()
	for _, route := range s.routes {
		path := route.path
		method := route.method
		router.HandleFunc(path, method)
	}
	return router
}

func (s *APIServer) InitMiddleware(router *http.ServeMux) http.Handler {
	if s.activeMiddlewares == nil {
		return router
	}

	var middlewares []Middleware
	for _, value := range s.activeMiddlewares {
		switch value {
		case "logging":
			logger := NewTableLogger(os.Stdout)
			middlewares = append(middlewares, LoggingMiddleware(logger))
		case "emoji":
			middlewares = append(middlewares, EmojiMiddleware)
		case "auth":
			middlewares = append(middlewares, RequireAuthMiddleware)
		}
	}

	if len(middlewares) == 0 {
		return router
	}

	middlewareChain := ChainMiddleware(middlewares...)
	return middlewareChain(router)
}

func (s *APIServer) Run() error {

	if len(s.routes) == 0 {
		return errors.New("No routes added to the server")
	}

	// Method for initializing the router and appending the routes to it
	router := s.InitRouter()

	// Initialize the middleware chain
	handler := s.InitMiddleware(router)

	// Versioning the API
	v1 := http.NewServeMux()
	v1.Handle("/api/v1", http.StripPrefix("/api/v1", router))

	server := http.Server{
		Addr:    s.addr,
		Handler: handler,
	}

	log.Printf("Starting server on %s", s.addr)
	return server.ListenAndServe()
}

// LoggingMiddleware logs incoming requests along with the assigned emoji.
func LoggingMiddleware(tl *TableLogger) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Retrieve the emoji from the context
			emoji, ok := r.Context().Value(emojiKey).(string)
			if !ok || emoji == "" {
				emoji = "🤷‍♂️" // Default shrugging man emoji if not found
			}

			// Log the request details along with the emoji
			tl.LogRequest(emoji, r.Method, r.URL.Path, r.RemoteAddr)

			// Proceed to the next middleware/handler
			next.ServeHTTP(w, r)
		}
	}
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

// EmojiMiddleware assigns an emoji based on the time of day.
// There's a 10% chance it assigns no time-based emoji, triggering the fallback.
func EmojiMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Seed the random number generator
		rand.Seed(time.Now().UnixNano())

		// 10% chance to skip time-based emoji assignment
		skip := rand.Intn(100) < 10 // 10% probability

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

type Middleware func(http.Handler) http.HandlerFunc

func ChainMiddleware(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}

		return next.ServeHTTP
	}
}
