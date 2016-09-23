package main

//go:generate go-bindata -o assets.go assets/

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/sontags/env"
)

type configuration struct {
	Listener string `json:"listener"`
	API      string `json:"api"`
	DB       string `json:"db"`
}

func (c configuration) String() string {
	b, _ := json.Marshal(c)
	return string(b)
}

var config configuration

func init() {
	env.Var(&config.Listener, "LISTENER", "0.0.0.0:8080", "Address/port to bind to")
	env.Var(&config.API, "API", "http://127.0.0.1:8080/api/", "Base URL where the API is reachable public")
	env.Var(&config.DB, "DB", "./shouts.db", "Path to database file")
}

var shouts Shouts

func main() {
	env.Parse("SHOUT", false)

	var err error
	shouts, err = NewShouts(config.DB)
	if err != nil {
		log.Fatalf("Error while opening/creating db: %s", err)
	}
	defer shouts.Close()

	r := mux.NewRouter()
	r.HandleFunc("/api/information/", listInformation).Methods("GET")
	r.HandleFunc("/api/information/", postInformation).Methods("POST")
	r.HandleFunc("/api/information/{id}/", getInformation).Methods("GET")
	r.HandleFunc("/api/information/{id}/", updateInformation).Methods("PUT")
	r.HandleFunc("/api/whoami/", whoAmI).Methods("GET")
	r.HandleFunc("/ui/", renderUI).Methods("GET")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/", 301)
	}).Methods("GET")

	chain := alice.New(
		RecoverPanic,
		CorsHeaders,
		UserContext,
	).Then(r)

	log.Fatal(http.ListenAndServe(config.Listener, chain))
}
