package main

import (
	"net/http"

	"github.com/CloudyKit/jet/v6"
)

func main() {
	// Create a new APIServer instance
	server := NewAPIServer(":8080")

	var views = jet.NewSet(
		jet.NewOSFileSystemLoader("./views"),
	)

	route01 := Route{
		path: "GET /users/{userId}",
		method: func(w http.ResponseWriter, r *http.Request) {
			userId := r.PathValue("userId")
			w.Write([]byte("Oi mate, yer Id is: " + userId))
		},
	}

	route02 := Route{
		path: "/",
		method: func(w http.ResponseWriter, r *http.Request) {
			indexView, err := views.GetTemplate("index.jet")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			indexView.Execute(w, nil, nil)
		},
	}

	useMiddlewares := []string{"logging"}

	// Add the route to the server
	server.AddRoute(route01)
	server.AddRoute(route02)
	server.AddMiddleware(useMiddlewares)

	server.Run()
}
