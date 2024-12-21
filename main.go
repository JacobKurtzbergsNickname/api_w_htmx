package main

import (
	"fmt"
	"net/http"

	"github.com/CloudyKit/jet/v6"
)

func main() {
	// Create a new APIServer instance
	server := NewAPIServer(":8080")

	creatures := GetAllEntities()
	fmt.Println(creatures)

	var views = jet.NewSet(
		jet.NewOSFileSystemLoader("./views"),
	)

	root := Route{
		path: "/",
		method: func(w http.ResponseWriter, r *http.Request) {
			indexView, err := views.GetTemplate("index.jet")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			indexView.Execute(w, nil, creatures)
		},
	}

	useMiddlewares := []string{"logging"}

	// Add the route to the server
	server.AddRoute(root)
	server.AddMiddleware(useMiddlewares)

	server.Run()
}
