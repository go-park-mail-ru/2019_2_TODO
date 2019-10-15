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

	siteMux := http.NewServeMux()

	siteMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		w.Write([]byte("{}"))
	})

	siteMux.HandleFunc("/signup/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		if r.Method == http.MethodPost {
			handlers.handleSignUp(w, r)
			return
		}

	})

	siteMux.HandleFunc("/signin/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		if r.Method == http.MethodPost {
			handlers.handleSignIn(w, r)
			return
		}

		handlers.handleSignInGet(w, r)
	})

	siteMux.HandleFunc("/signin/profile/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		if r.Method == http.MethodPost {
			handlers.handleChangeProfile(w, r)
			return
		}

		handlers.handleGetProfile(w, r)

	})

	siteMux.HandleFunc("/signin/profileImage/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		if r.Method == http.MethodPost {
			handlers.handleChangeImage(w, r)
			return
		}

	})

	siteMux.HandleFunc("/logout/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		handlers.handleLogout(w, r)
	})

	siteMux.HandleFunc("/checkUsers/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application-json")

		handlers.checkUsersForTesting(w, r)
	})

	siteMux.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")

		avatar := handlers.ReadCookieAvatar(w, r)

		log.Println(avatar)

		http.ServeFile(w, r, "/root/golang/test/2019_2_TODO/server/"+avatar)
	})

	siteHandler := corsMiddware(siteMux)
	siteHandler = panicMiddware(siteHandler)
	siteHandler = accessLogMiddware(siteHandler)

	http.ListenAndServe(":8080", siteHandler)
}
