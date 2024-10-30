package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type Application struct {
	DB      *sql.DB
	Router  *mux.Router
	Store   sessions.Store
	UserSvc UserService
}

func NewApplication() (*Application, error) {
	db, err := sql.Open("mysql", "root:password@tcp(localhost:3306)/crud_db?parseTime=true")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	app := &Application{
		DB:      db,
		Router:  mux.NewRouter(),
		Store:   sessions.NewCookieStore([]byte("secret-key")),
		UserSvc: NewUserService(db),
	}

	app.routes()
	return app, nil
}

func (app *Application) routes() {
	app.Router.HandleFunc("/register", app.registerHandler).Methods("POST")
	app.Router.HandleFunc("/login", app.loginHandler).Methods("POST")
	app.Router.HandleFunc("/users", app.authMiddleware(app.listUsersHandler)).Methods("GET")
	app.Router.HandleFunc("/users/{id}", app.authMiddleware(app.deleteUserHandler)).Methods("DELETE")
}
