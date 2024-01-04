package main

import (
	"encoding/json"
	"net/http"
	"net/mail"

	"github.com/aliasboink/go_web_server/internal/database"
)

func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func handlerPostUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	if !validEmail(params.Email) {
		respondWithError(w, 422, "Invalid email!")
		return
	}
	db, err := database.NewDB("database.json")
	if err != nil {
		respondWithError(w, 500, "Something went wrong with the DB!")
		return
	}
	user, err := db.CreateUser(params.Email)
	if err != nil {
		respondWithError(w, 500, "Something went wrong creating the user!")
		return
	}
	respondWithJSON(w, 201, user)
}
