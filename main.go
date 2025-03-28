package main

import (
	"awesomeProject/v1"
	"log"
	"net/http"
)

func main() {
	app, err := v1.NewApplication()
	if err != nil {
		log.Fatal(err)
	}
	defer app.DB.Close()

	log.Printf("Server starting on port 8080\n")
	log.Fatal(http.ListenAndServe(":8080", app.Router))
}
