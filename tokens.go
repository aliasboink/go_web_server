package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aliasboink/go_web_server/internal/database"
	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlerPostRevoke(w http.ResponseWriter, r *http.Request) {
	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong with the DB!")
		return
	}
	err = db.RevokeToken(tokenString)
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong with the DB!")
		return
	}
	w.WriteHeader(200)
	return
}

func (cfg *apiConfig) handlerPostRefresh(w http.ResponseWriter, r *http.Request) {
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
	if err != nil || issuer != "Chirpy-Refresh" {
		log.Print(err.Error())
		respondWithError(w, 401, "Unauthorized!")
		return
	}
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong with the DB!")
		return
	}
	err = db.CheckRevocation(tokenString)
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 401, "Unauthorized!")
		return
	}
	// Create Access JWT Token
	userId, err := jwtToken.Claims.GetSubject()
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	now := time.Now()
	jwtClaimsAccess := jwt.RegisteredClaims{
		Issuer:    "Chirpy-Access",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		Subject:   userId,
	}
	jwtAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaimsAccess)
	// https://github.com/dgrijalva/jwt-go/issues/65#issuecomment-98019456
	jwtStringAccess, err := jwtAccessToken.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	response := struct {
		Token string `json:"token"`
	}{
		Token: jwtStringAccess,
	}
	respondWithJSON(w, 200, response)
	return
}
