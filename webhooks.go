package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/aliasboink/go_web_server/internal/database"
)

func (cfg *apiConfig) handlerPostPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	polkaSecret := strings.TrimPrefix(r.Header.Get("Authorization"), "ApiKey ")
	if polkaSecret != cfg.polkaSecret {
		w.WriteHeader(401)
		return
	}
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserId int `json:"user_id"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	if params.Event != "user.upgraded" {
		w.WriteHeader(200)
		return
	}
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	_, err = db.UpgradeUser(params.Data.UserId)
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	w.WriteHeader(200)
	return
}
