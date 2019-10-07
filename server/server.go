package main

import (
	"log"
	"net/http"
	"sync"
)

func main() {
	handlers := Handlers{
		users: make([]Credentials, 0),
		mu:    &sync.Mutex{},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		log.Println(r.URL.Path)

		w.Write([]byte("{}"))
	})

	http.HandleFunc("/signup/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		log.Println(r.URL.Path)

		if r.Method == http.MethodPost {
			handlers.handleSignUp(w, r)
			return
		}

	})

	http.HandleFunc("/signin/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		log.Println(r.URL.Path)

		if r.Method == http.MethodPost {
			handlers.handleSignIn(w, r)
			return
		}

		handlers.handleSignInGet(w, r)
	})

	http.HandleFunc("/signin/profile/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		log.Println(r.URL.Path)

		if r.Method == http.MethodPost {
			handlers.handleChangeProfile(w, r)
			return
		}

		handlers.handleGetProfile(w, r)

	})

	http.HandleFunc("/signin/profileImage/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		log.Println(r.URL.Path)

		if r.Method == http.MethodPost {
			handlers.handleChangeImage(w, r)
			return
		}

	})

	http.HandleFunc("/logout/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		log.Println(r.URL.Path)

		handlers.handleLogout(w, r)
	})

	http.HandleFunc("/checkUsers/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		log.Println(r.URL.Path)

		handlers.checkUsersForTesting(w, r)
	})

	http.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", frontIp)
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		log.Println(r.URL.Path)

		absPath := "http://93.171.139.196:780/"
		avatar := handlers.ReadCookieAvatar(w, r)

		http.ServeFile(w, r, absPath+avatar)
	})

	http.ListenAndServe(":8080", nil)
}
