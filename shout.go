package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

const bucketName = "shouts"

var shouts Shouts

type Shout struct {
	ID uint64    `json:"id"`
	U  string    `json:"user"`
	M  string    `json:"message"`
	D  time.Time `json:"date"`
}

func (s Shout) Serialize() []byte {
	out, _ := json.Marshal(s)
	return out
}

func (s *Shout) Init(in []byte) error {
	return json.Unmarshal(in, s)
}

type Shouts struct {
	db  *bolt.DB
	loc string
}

func NewShouts(location string) (Shouts, error) {
	s := Shouts{
		loc: location,
	}

	// open database
	var err error
	s.db, err = bolt.Open(s.loc, 0600, nil)
	if err != nil {
		return s, err
	}

	// create required buckets
	s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return s, err
	}

	return s, nil
}

func (s Shouts) Close() {
	s.db.Close()
}

func (s Shouts) Save(user string, message string) (uint64, error) {
	shout := Shout{
		U: user,
		M: message,
		D: time.Now(),
	}
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		shout.ID, _ = b.NextSequence()
		idb := itob(shout.ID)
		return b.Put(idb, shout.Serialize())
	})
	return shout.ID, err
}

func (s Shouts) Update(id uint64, user string, message string) (uint64, error) {
	shout := Shout{
		ID: id,
		U:  user,
		M:  message,
		D:  time.Now(),
	}
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		idb := itob(shout.ID)
		return b.Put(idb, shout.Serialize())
	})
	return shout.ID, err
}

func (s Shouts) List() []uint64 {
	var out []uint64
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			out = append(out, btoi(k))
		}
		return nil
	})
	return out
}

func (s Shouts) Get(id uint64) (Shout, error) {
	shout := Shout{
		ID: 0,
	}
	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return shout.Init(b.Get(itob(id)))
	})

	if shout.ID == 0 {
		return shout, errors.New("shout does not exist")
	}
	return shout, nil
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func btoi(v []byte) uint64 {
	return binary.BigEndian.Uint64(v)
}

func getBodyAsBytes(body io.ReadCloser) ([]byte, error) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}

func listInformation(w http.ResponseWriter, r *http.Request) {
	list := shouts.List()

	out, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error")
		return
	}
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
	fmt.Fprintf(w, "%s", out)
}

func postInformation(w http.ResponseWriter, r *http.Request) {
	message, _ := getBodyAsBytes(r.Body)
	shoutId, err := shouts.Save("aoeu", string(message))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error while saving to DB")
		return
	}
	fmt.Fprintf(w, "%d", shoutId)
}

func updateInformation(w http.ResponseWriter, r *http.Request) {
	message, _ := getBodyAsBytes(r.Body)

	vars := mux.Vars(r)
	id := vars["id"]
	shoutId, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error while parsing ID")
		return
	}

	shoutId, err = shouts.Update(shoutId, "aoeu", string(message))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal Server Error while saving to DB")
		return
	}
	fmt.Fprintf(w, "%d", shoutId)
}

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

	log.Fatal(http.ListenAndServe(":8080", r))
}
