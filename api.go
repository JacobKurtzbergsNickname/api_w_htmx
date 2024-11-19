package main

import (
	"errors"
	"log"
	"net/http"
	"os"
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
