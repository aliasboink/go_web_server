package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

type Chirp struct {
	Id   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	mux := sync.RWMutex{}
	db := DB{path: path, mux: &mux}
	err := db.ensureDB()
	if err != nil {
		return &DB{}, err
	}
	return &db, nil
}

// CreateChirp creates a new chirp and saves it to disk
// Assume ID's are in order
func (db *DB) CreateChirp(body string) (Chirp, error) {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return Chirp{}, err
	}
	var newChirpId int
	if len(dbStructure.Chirps) < 1 {
		newChirpId = 1
	} else {
		lastChirp := dbStructure.Chirps[len(dbStructure.Chirps)]
		newChirpId = lastChirp.Id + 1
	}
	newChirp := Chirp{
		Id:   newChirpId,
		Body: body,
	}
	dbStructure.Chirps[newChirp.Id] = newChirp
	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}
	return newChirp, nil
}

// Nearly identical to CreateChirp
func (db *DB) CreateUser(email string) (User, error) {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return User{}, err
	}
	var newUserId int
	if len(dbStructure.Users) < 1 {
		newUserId = 1
	} else {
		lastUser := dbStructure.Users[len(dbStructure.Users)]
		newUserId = lastUser.Id + 1
	}
	newUser := User{
		Id:    newUserId,
		Email: email,
	}
	dbStructure.Users[newUser.Id] = newUser
	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}
	return newUser, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.LoadDB()
	if err != nil {
		return []Chirp{}, err
	}
	chirps := make([]Chirp, len(dbStructure.Chirps))
	for index, chirp := range dbStructure.Chirps {
		chirps[index-1] = chirp
	}
	return chirps, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	_, err := os.Stat(db.path)
	// I don't really like this bit of code frankly
	if errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(db.path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = file.Write([]byte("{}"))
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

// loadDB reads the database file into memory
func (db *DB) LoadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	databaseBytes, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}
	if jsonInfo, err := os.Stat(db.path); err != nil {
		return DBStructure{}, err
	} else if jsonInfo.Size() <= 2 {
		return DBStructure{make(map[int]Chirp), make(map[int]User)}, nil
	}
	var dbStructure DBStructure
	err = json.Unmarshal(databaseBytes, &dbStructure)
	if err != nil {
		log.Printf("Error unmarshalling JSON: %s", err)
		return DBStructure{}, err
	}
	return dbStructure, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()
	dbStructureBytes, err := json.MarshalIndent(dbStructure, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, dbStructureBytes, 0666)
	if err != nil {
		return err
	}
	return nil
}
