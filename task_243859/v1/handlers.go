package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (app *Application) registerHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := &User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}

	if err := app.UserSvc.Create(user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (app *Application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := app.UserSvc.Authenticate(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	session, _ := app.Store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Values["user_id"] = user.ID
	session.Save(r, w)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (app *Application) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := app.UserSvc.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func (app *Application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := app.UserSvc.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *Application) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := app.Store.Get(r, "session-name")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
