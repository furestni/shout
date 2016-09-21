package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func listInformation(w http.ResponseWriter, r *http.Request) {
	list := shouts.List()

	out, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "%s", out)
}

func getInformation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	shoutId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error while parsing ID")
		return
	}
	shout, err := shouts.Get(shoutId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error while fetching from DB")
		return
	}
	out := shout.Serialize()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error while rendering to json")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "%s", out)
}

func postInformation(w http.ResponseWriter, r *http.Request) {
	message, _ := getBodyAsBytes(r.Body)
	user := getUserName(r)

	shoutId, err := shouts.Save(user, string(message))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error while saving to DB")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "%d", shoutId)
}

func updateInformation(w http.ResponseWriter, r *http.Request) {
	message, _ := getBodyAsBytes(r.Body)
	user := getUserName(r)

	vars := mux.Vars(r)
	id := vars["id"]
	shoutId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error while parsing ID")
		return
	}

	shoutId, err = shouts.Update(shoutId, user, string(message))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error while saving to DB")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "%d", shoutId)
}

func whoAmI(w http.ResponseWriter, r *http.Request) {
	user := getUserName(r)
	fmt.Fprintf(w, "%s", user)
}

func getBodyAsBytes(body io.ReadCloser) ([]byte, error) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}

func getUserName(r *http.Request) string {
	return r.Header.Get(HeaderUserName)
}
