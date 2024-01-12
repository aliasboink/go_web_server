package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/aliasboink/go_web_server/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	// This could be a function
	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims := jwt.RegisteredClaims{}
	jwtToken, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.jwtSecret), nil
	})
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 401, "Unauthorized!")
		return
	}
	issuer, err := claims.GetIssuer()
	if err != nil || issuer != "Chirpy-Access" {
		respondWithError(w, 401, "Unauthorized!")
		return
	}
	// Up to here
	userId, err := jwtToken.Claims.GetSubject()
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long!")
		return
	}
	db, err := database.NewDB("database.json")
	if err != nil {
		respondWithError(w, 500, "Something went wrong with the DB!")
		return
	}
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	chirp, err := db.CreateChirp(cleanTheProfanities(params.Body, profaneWords), userId)
	if err != nil {
		respondWithError(w, 500, "Something went wrong creating the chirp!")
		return
	}
	respondWithJSON(w, 201, chirp)
}

func handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	db, err := database.NewDB("database.json")
	if err != nil {
		respondWithError(w, 500, "Something went wrong with the DB!")
		return
	}
	chirps, err := db.GetChirps()
	if err != nil {
		respondWithError(w, 500, "Something went wrong with chirps!")
		return
	}
	respondWithJSON(w, 200, chirps)
}

func handlerGetChirpWithId(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	db, err := database.NewDB("database.json")
	if err != nil {
		respondWithError(w, 500, "Something went wrong with the DB!")
		return
	}
	dbStructure, err := db.LoadDB()
	if err != nil {
		respondWithError(w, 500, "Something went wrong with the DB loading!")
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	if chirp, ok := dbStructure.Chirps[idInt]; ok {
		respondWithJSON(w, 200, chirp)
		return
	} else {
		respondWithError(w, 404, "Not found!")
	}
}
