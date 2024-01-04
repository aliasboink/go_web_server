package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aliasboink/go_web_server/internal/database"
	"github.com/go-chi/chi/v5"
)

// func handlerValidateChirps(w http.ResponseWriter, r *http.Request) {

// 	type parameters struct {
// 		Body string `json:"body"`
// 	}

// 	decoder := json.NewDecoder(r.Body)
// 	params := parameters{}
// 	err := decoder.Decode(&params)
// 	if err != nil {
// 		respondWithError(w, 500, "Something went wrong")
// 		return
// 	}

// 	if len(params.Body) > 140 {
// 		respondWithError(w, 400, "Chirp is too long")
// 		return
// 	}

// 	type returnVals struct {
// 		Valid       bool   `json:"valid"`
// 		CleanedBody string `json:"claned_body"`
// 	}
// 	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
// 	respBody := returnVals{
// 		Valid:       true,
// 		CleanedBody: cleanTheProfanities(params.Body, profaneWords),
// 	}
// 	respondWithJSON(w, 200, respBody)
// }

func handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
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
	chirp, err := db.CreateChirp(cleanTheProfanities(params.Body, profaneWords))
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
