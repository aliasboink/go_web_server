package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/aliasboink/go_web_server/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (cfg *apiConfig) handlerPostLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
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
	dbStructure, err := db.LoadDB()
	for _, user := range dbStructure.Users {
		if strings.ToLower(user.Email) == strings.ToLower(params.Email) {
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(params.Password))
			if err != nil {
				respondWithError(w, 401, "Wrong password!")
				return
			}
			// Create JWT Token
			now := time.Now()
			if params.ExpiresInSeconds == 0 || params.ExpiresInSeconds > 86400 {
				params.ExpiresInSeconds = 86400
			}
			jwtClaims := jwt.RegisteredClaims{
				Issuer:    "Chirpy",
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Second * time.Duration(params.ExpiresInSeconds))),
				Subject:   fmt.Sprintf("%d", user.Id),
			}

			jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
			// interface{}, but the string has to be cast to []byte, ok...
			// https://github.com/dgrijalva/jwt-go/issues/65#issuecomment-98019456
			jwtString, err := jwtToken.SignedString([]byte(cfg.jwtSecret))
			if err != nil {
				log.Print(err.Error())
				respondWithError(w, 500, "Something went wrong!")
				return
			}
			response := struct {
				Id    int    `json:"id"`
				Email string `json:"email"`
				Token string `json:"token,omitempty"`
			}{
				Id:    user.Id,
				Email: user.Email,
				Token: jwtString,
			}
			respondWithJSON(w, 200, response)
			return
		}
	}
	respondWithError(w, 401, "Wrong email!")
	return
}

func (cfg *apiConfig) handlerPutUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong!")
		return
	}
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
	userId, err := jwtToken.Claims.GetSubject()
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	db, err := database.NewDB("database.json")
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	user, err := db.UpdateUser(userIdInt, params.Email, params.Password)
	if err != nil {
		log.Print(err.Error())
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	respondWithJSON(w, 200, user)
	return
}

func handlerPostUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), 10)
	if err != nil {
		respondWithError(w, 500, "Something went wrong!")
		return
	}
	user, err := db.CreateUser(params.Email, fmt.Sprintf("%s", hashedPassword))
	if err != nil {
		// Is this okay?
		respondWithError(w, 500, err.Error())
		return
	}
	response := struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
	}{
		Id:    user.Id,
		Email: user.Email,
	}
	respondWithJSON(w, 201, response)
}
