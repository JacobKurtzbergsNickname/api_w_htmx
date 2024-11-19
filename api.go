package main

import (
	"log"
	"net/http"
)

type APIServer struct {
	addr string
}

func NewAPIServer(addr string) *APIServer {
	return &APIServer{addr: addr}
}

func (s *APIServer) Run() error {
	router := http.NewServeMux()

	router.HandleFunc("GET /users/{userId}", func(w http.ResponseWriter, r *http.Request) {
		// Extract the userId from the URL
		userId := r.PathValue("userId")
		w.Write([]byte("Oi mate, yer Id is: " + userId))
	})

	v1 := http.NewServeMux()
	v1.Handle("/api/v1", http.StripPrefix("/api/v1", router))

	// Logger first, otherwise we won't see the log for unauthorized requests
	middlewareChain := ChainMiddleware(
		RequestLoggerMiddleware,
		RequireAuthMiddleware,
	)

	server := http.Server{
		Addr:    s.addr,
		Handler: middlewareChain(router),
	}

	log.Printf("Starting server on %s", s.addr)
	return server.ListenAndServe()
}

func RequestLoggerMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s Request from %s for %s", r.Method, r.RemoteAddr, r.URL)
		next.ServeHTTP(w, r)
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

type Middleware func(http.Handler) http.HandlerFunc

func ChainMiddleware(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}

		return next.ServeHTTP
	}
}
