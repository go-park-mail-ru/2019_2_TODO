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
		SetCookie(w, "Nickname")
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
			SetCookie(w, "Hello")
			return
		}
	})

	http.HandleFunc("/profile/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		log.Println(r.URL.Path)

		if r.Method == http.MethodPost {
			handlers.handleChangeProfile(w, r)
			return
		}
		handlers.handleGetProfile(w, r)
	})

	http.HandleFunc("/logout/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		log.Println(r.URL.Path)

		handlers.handleLogout(w, r)
	})

	http.ListenAndServe(":8080", nil)
}
