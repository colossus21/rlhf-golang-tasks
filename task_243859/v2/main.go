
package main

import (
    "log"
)

func main() {
    app, err := NewApplication()
    if err != nil {
        log.Fatal(err)
    }
    defer app.DB.Close()

    log.Printf("Server starting on port 8080")
    if err := app.Router.Run(":8080"); err != nil {
        log.Fatal(err)
    }
}
