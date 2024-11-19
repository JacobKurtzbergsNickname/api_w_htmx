package main

import "net/http"

func main() {
	// Create a new APIServer instance
	server := NewAPIServer(":8080")

	route := Route{
		path: "GET /users/{userId}",
		method: func(w http.ResponseWriter, r *http.Request) {
			userId := r.PathValue("userId")
			w.Write([]byte("Oi mate, yer Id is: " + userId))
		},
	}

	useMiddlewares := []string{"logging", "auth"}

	// Add the route to the server
	server.AddRoute(route)
	server.AddMiddleware(useMiddlewares)

	server.Run()
}
