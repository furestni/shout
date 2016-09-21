package main

import (
	"encoding/json"
	"time"
)

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
