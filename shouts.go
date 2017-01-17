package main

import (
	"encoding/binary"
	"errors"
	"time"

	"github.com/boltdb/bolt"
)

const bucketName = "shouts"

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

func (s Shouts) Delete(id uint64) (uint64, error) {
        err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
                idb := itob(id)
                return b.Delete(idb)
	})
        return id, err
}


func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func btoi(v []byte) uint64 {
	return binary.BigEndian.Uint64(v)
}
