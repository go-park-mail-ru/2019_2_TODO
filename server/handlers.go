package main

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/securecookie"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)

type CredentialsInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Image    string `json:"image"`
}

type Credentials struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Image    string `json:"image"`
}

type Handlers struct {
	users []Credentials
	mu    *sync.Mutex
}

func (h *Handlers) handleSignUp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		return
	}

	decoder := json.NewDecoder(r.Body)
	newUserInput := new(CredentialsInput)

	err := decoder.Decode(newUserInput)
	if err != nil {
		log.Printf("Error while decoding body: %s", err)
		w.Write([]byte("{}"))
		return
	}

	h.mu.Lock()

	var idUser uint64 = 0
	var defaultImage = "images/avatar.png"

	if len(h.users) > 0 {
		idUser = h.users[len(h.users)-1].ID + 1
	}

	h.users = append(h.users, Credentials{
		ID:       idUser,
		Username: newUserInput.Username,
		Password: newUserInput.Password,
		Image:    defaultImage,
	})
	h.mu.Unlock()

	SetCookie(w, newUserInput.Username)

	http.Redirect(w, r, "http://localhost:8080", http.StatusSeeOther)
}

func (h *Handlers) handleSignIn(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		return
	}

	decoder := json.NewDecoder(r.Body)
	authCredentials := new(CredentialsInput)

	err := decoder.Decode(authCredentials)
	if err != nil {
		log.Printf("Error while decoding body: %s", err)
		w.Write([]byte("{}"))
		return
	}

	h.mu.Lock()

	accounts := h.users

	err = h.checkUsername(accounts, authCredentials)

	if err != nil {
		log.Printf("%s", err)
		w.Write([]byte("{}"))
		return
	}

	err = h.checkPassword(accounts, authCredentials)

	if err != nil {
		log.Printf("%s", err)
		w.Write([]byte("{}"))
		return
	}

	SetCookie(w, authCredentials.Username)

	h.mu.Unlock()

	http.Redirect(w, r, "http://localhost:8080", http.StatusSeeOther)

}

func (h *Handlers) handleChangeProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		return
	}

	if h.ReadCookieUsername(w, r) == "" {
		http.Redirect(w, r, "http://localhost:8080/signin/", http.StatusSeeOther)
		return
	}

	decoder := json.NewDecoder(r.Body)
	changeProfileCredentials := new(CredentialsInput)

	err := decoder.Decode(changeProfileCredentials)
	if err != nil {
		log.Printf("Error while decoding body: %s", err)
		w.Write([]byte("{}"))
		return
	}

	h.mu.Lock()

	accounts := h.users

	h.changeProfile(accounts, changeProfileCredentials)

	if changeProfileCredentials.Username != "" {
		loadAvatar(w, r, changeProfileCredentials.Username)
		changeProfileCredentials.Image = "images/" + changeProfileCredentials.Username + ".jpg"
		ClearCookie(w)
		SetCookie(w, changeProfileCredentials.Username)
	} else {
		loadAvatar(w, r, h.ReadCookieUsername(w, r))
		changeProfileCredentials.Image = "images/" + changeProfileCredentials.Username + ".jpg"
	}

	h.mu.Unlock()

}

func loadAvatar(w http.ResponseWriter, r *http.Request, username string) {
	src, hdr, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer src.Close()

	dst, err := os.Create(filepath.Join("/home/toringol/2019_2_TODO/server/images/", hdr.Filename))
	os.Rename(filepath.Join("/home/toringol/2019_2_TODO/server/images/", hdr.Filename),
		filepath.Join("/home/toringol/2019_2_TODO/server/images/", username+".jpg"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer dst.Close()

	io.Copy(dst, src)
}

func (h *Handlers) handleGetProfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	encoder := json.NewEncoder(w)

	h.mu.Lock()
	err := encoder.Encode(h.ReadCookieUsername(w, r))
	err1 := encoder.Encode(h.ReadCookieAvatar(w, r))
	h.mu.Unlock()

	if err != nil || err1 != nil {
		log.Printf("Error while encoding json: %s", err)
		w.Write([]byte("{}"))
	}
}

func (h *Handlers) handleLogout(w http.ResponseWriter, r *http.Request) {
	ClearCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) checkUsersForTesting(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	encoder := json.NewEncoder(w)
	h.mu.Lock()
	err := encoder.Encode(h.users)
	h.mu.Unlock()
	if err != nil {
		log.Printf("error while marshalling JSON: %s", err)
		w.Write([]byte("{}"))
		return
	}
	log.Println(h.users)
}

func (h *Handlers) checkUsername(accounts []Credentials, authCredentials *CredentialsInput) error {
	sort.Slice(accounts[:], func(i, j int) bool {
		return accounts[i].Username < accounts[j].Username
	})

	iter := sort.Search(len(accounts), func(i int) bool {
		return accounts[i].Username == authCredentials.Username
	})

	if iter >= len(accounts) || accounts[iter].Username != authCredentials.Username {
		return errors.New("No such user")
	}
	return nil
}

func (h *Handlers) checkPassword(accounts []Credentials, authCredentials *CredentialsInput) error {
	sort.Slice(accounts[:], func(i, j int) bool {
		return accounts[i].Password < accounts[j].Password
	})

	iter := sort.Search(len(accounts), func(i int) bool {
		return accounts[i].Password == authCredentials.Password
	})

	if iter >= len(accounts) || accounts[iter].Password != authCredentials.Password {
		return errors.New("Wrong password")
	}
	return nil
}

func (h *Handlers) changeProfile(accounts []Credentials, changeProfileCredentials *CredentialsInput) {
	sort.Slice(accounts[:], func(i, j int) bool {
		return accounts[i].Username < accounts[j].Username
	})

	iter := sort.Search(len(accounts), func(i int) bool {
		return accounts[i].Username == changeProfileCredentials.Username
	})

	accounts[iter].Username = changeProfileCredentials.Username
	accounts[iter].Password = changeProfileCredentials.Password
}

func SetCookie(w http.ResponseWriter, userName string) {
	value := map[string]string{
		"username": userName,
	}

	if encoded, err := cookieHandler.Encode("session_token", value); err == nil {
		expiration := time.Now().Add(24 * time.Hour)
		cookie := http.Cookie{
			Name:    "session_token",
			Value:   encoded,
			Path:    "/",
			Expires: expiration,
		}

		http.SetCookie(w, &cookie)
	}
}

func ClearCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(w, cookie)
}

func (h *Handlers) ReadCookieUsername(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie("session_token"); err == nil {
		value := make(map[string]string)
		if err = cookieHandler.Decode("session_token", cookie.Value, &value); err == nil {
			return value["username"]
		}
	}
	return ""
}

func (h *Handlers) ReadCookieAvatar(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie("session_token"); err == nil {
		value := make(map[string]string)
		if err = cookieHandler.Decode("session_token", cookie.Value, &value); err == nil {
			accounts := h.users
			sort.Slice(accounts[:], func(i, j int) bool {
				return accounts[i].Username < accounts[j].Username
			})

			iter := sort.Search(len(accounts), func(i int) bool {
				return accounts[i].Username == value["username"]
			})
			return accounts[iter].Image
		}
	}
	return ""
}
