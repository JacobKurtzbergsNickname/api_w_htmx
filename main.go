package main

func main() {
	// Create a new APIServer instance
	server := NewAPIServer(":8080")
	server.Run()
}
