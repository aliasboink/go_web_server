package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

func deleteDatabase(path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

func outputHTML(w http.ResponseWriter, filename string, data interface{}) {
	t, err := template.ParseFiles(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type returnVals struct {
		Error string `json:"error"`
	}
	respBody := returnVals{
		Error: msg,
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
	return
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
	return
}

func cleanTheProfanities(msg string, profaneWords []string) string {
	splitMsg := strings.Split(msg, " ")
	var msgCensoredSlice []string = make([]string, len(splitMsg))
	var found bool
	for index, word := range splitMsg {
		found = false
		for _, profaneWord := range profaneWords {
			if strings.ToLower(word) == profaneWord {
				msgCensoredSlice[index] = "****"
				found = true
				break
			}
		}
		if found == false {
			msgCensoredSlice[index] = word
		}
	}
	return strings.Join(msgCensoredSlice, " ")
}
