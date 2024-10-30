package main

import (
	"log"
	"net/http"
)

func main() {
	app, err := NewApplication()
	if err != nil {
		log.Fatal(err)
	}
	defer app.DB.Close()

	log.Printf("Server starting on port 8080\n")
	log.Fatal(http.ListenAndServe(":8080", app.Router))
}
