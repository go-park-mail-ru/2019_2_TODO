package main

import (
	"encoding/json"
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

		w.Header().Set("Access-Control-Allow-Origin", clientIp)
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		cookieUsername := handlers.ReadCookieUsername(w, r)

		log.Println(cookieUsername)

		if cookieUsername != "" {
			cookieUsernameInput := CredentialsInput{
				Username: cookieUsername,
			}

			encoder := json.NewEncoder(w)
			err := encoder.Encode(cookieUsernameInput)
			if err != nil {
				log.Println("Error while encoding")
				w.Write([]byte("{}"))
				return
			}
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

	http.HandleFunc("/profileImage/", func(w http.ResponseWriter, r *http.Request) {
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

	http.ListenAndServe(":80", nil)
}
