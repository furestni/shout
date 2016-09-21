package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func CorsHeaders(next http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		if origin := req.Header.Get("Origin"); origin != "" {
			res.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			res.Header().Set("Access-Control-Allow-Origin", "*")
		}
		res.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		res.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Accept, Authorization")
		res.Header().Set("Access-Control-Allow-Credentials", "true")
		// if req.Method == http.MethodOptions {
		if req.Method == "OPTIONS" {
			fmt.Fprintf(res, "Hello")
			return
		}
		next.ServeHTTP(res, req)
	}

	return http.HandlerFunc(fn)
}

// RecoverPanic return an Internal Server Error if a panic occures during
// a handler call and recovers.
func RecoverPanic(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// HeaderUserName holds the string of the header field to store the username
const HeaderUserName = "S-UserName"

const (
	defaultUserName       = "Gordon Ramsey"
	defaultBrokenUserName = "Rordon Gamsey"
)

// UserContext reads the 'Authorization' header from the request, decodes the
// credentials and stores the user name as new header 'G-UserName'
func UserContext(next http.Handler) http.Handler {
	fn := func(res http.ResponseWriter, req *http.Request) {
		auth := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
		if len(auth) != 2 {
			req.Header.Set(HeaderUserName, defaultUserName)
		} else {
			username, _ := extractUserName(auth[1])
			req.Header.Set(HeaderUserName, username)
		}
		next.ServeHTTP(res, req)
	}

	return http.HandlerFunc(fn)
}

func extractUserName(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return defaultBrokenUserName, errors.New("Credentials are not properly encoded")
	}

	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return defaultBrokenUserName, errors.New("Decoded credentials are malformated")
	}

	return credentials[0], nil
}
