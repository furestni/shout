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
	env.Var(&config.Listener, "LISTENER", "0.0.0.0:8088", "Address/port to bind to")
	env.Var(&config.API, "API", "http://dashing.stxt.media.int:8088/api/", "Base URL where the API is reachable public")
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
        r.HandleFunc("/api/information/{id}/", deleteInformation).Methods("DELETE")
	r.HandleFunc("/api/whoami/", whoAmI).Methods("GET")
	r.HandleFunc("/ui/", renderUI).Methods("GET")

	chain := alice.New(
		RecoverPanic,
		CorsHeaders,
		UserContext,
	).Then(r)

	log.Fatal(http.ListenAndServe(config.Listener, chain))
}
