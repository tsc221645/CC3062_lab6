package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Match struct {
	ID        int    `json:"id"`
	HomeTeam  string `json:"homeTeam"`
	AwayTeam  string `json:"awayTeam"`
	MatchDate string `json:"matchDate"`
}

var db *sql.DB // database connection

// main - main function
func main() {
	var err error
	// open database
	db, err = sql.Open("sqlite3", "/data/matches.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable()

	router := mux.NewRouter()
	router.Use(middlewareCORS)
	// routes
	router.HandleFunc("/api/matches", getAllMatches).Methods("GET")
	router.HandleFunc("/api/matches/{id}", getMatchByID).Methods("GET")
	router.HandleFunc("/api/matches", createMatch).Methods("POST")
	router.HandleFunc("/api/matches/{id}", updateMatch).Methods("PUT")
	router.HandleFunc("/api/matches/{id}", deleteMatch).Methods("DELETE")
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./public"))))

	log.Println("Servidor escuchando en http://localhost:8080") //localhost:8080
	log.Fatal(http.ListenAndServe(":8080", router))
}

// middlewareCORS - middleware to handle CORS
func middlewareCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}

// createTable - creates the table if it does not exist
func createTable() {
	// create table
	query := `
	CREATE TABLE IF NOT EXISTS matches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		home_team TEXT NOT NULL,
		away_team TEXT NOT NULL,
		match_date TEXT NOT NULL
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// getAllMatches - returns all matches (GET)
func getAllMatches(w http.ResponseWriter, r *http.Request) {
	// query database
	rows, err := db.Query("SELECT id, home_team, away_team, match_date FROM matches")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	matches := []Match{}

	// iterate over rows
	for rows.Next() {
		var m Match
		err := rows.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		matches = append(matches, m)
	}

	w.Header().Set("Content-Type", "application/json") // set the response content type
	json.NewEncoder(w).Encode(matches)
}

// getMatchByID - returns a match by ID (GET)
func getMatchByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	// query database
	row := db.QueryRow("SELECT id, home_team, away_team, match_date FROM matches WHERE id = ?", id)

	var m Match
	err := row.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// createMatch - creates a new match (POST)
func createMatch(w http.ResponseWriter, r *http.Request) {
	var m Match
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// insert into database
	result, err := db.Exec("INSERT INTO matches (home_team, away_team, match_date) VALUES (?, ?, ?)", m.HomeTeam, m.AwayTeam, m.MatchDate)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	id, _ := result.LastInsertId()
	m.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// updateMatch - updates a match (PUT)
func updateMatch(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var m Match
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	// update database
	_, err = db.Exec("UPDATE matches SET home_team = ?, away_team = ?, match_date = ? WHERE id = ?", m.HomeTeam, m.AwayTeam, m.MatchDate, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// deleteMatch - deletes a match (DELETE)
func deleteMatch(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	// delete from database
	_, err := db.Exec("DELETE FROM matches WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
