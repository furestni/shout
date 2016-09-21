package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

var shouts Shouts

func main() {
	var err error
	shouts, err = NewShouts("./shouts.db")
	if err != nil {
		log.Fatalf("Error while opening/creating db: %s", err)
	}
	defer shouts.Close()

	r := mux.NewRouter()
	r.HandleFunc("/information/", listInformation).Methods("GET")
	r.HandleFunc("/information/", postInformation).Methods("POST")
	r.HandleFunc("/information/{id}/", getInformation).Methods("GET")
	r.HandleFunc("/information/{id}/", updateInformation).Methods("PUT")
	r.HandleFunc("/whoami/", whoAmI).Methods("GET")

	chain := alice.New(
		RecoverPanic,
		CorsHeaders,
		UserContext,
	).Then(r)

	log.Fatal(http.ListenAndServe(":8080", chain))
}
