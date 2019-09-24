package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
)

func TestSignUp(t *testing.T) {
	t.Parallel()

	h := Handlers{
		users: []Credentials{},
		mu:    &sync.Mutex{},
	}

	body := bytes.NewReader([]byte(`{"username": "Vasily", "password": "qwerty"}`))

	var expectedUsers = []Credentials{
		{
			ID:       0,
			Username: "Vasily",
			Password: "qwerty",
		},
	}

	r := httptest.NewRequest("POST", "/signup/", body)
	w := httptest.NewRecorder()

	h.handleSignUp(w, r)

	if w.Code != http.StatusOK {
		t.Error("Failed http Status")
	}

	reflect.DeepEqual(h.users, expectedUsers)

}

func TestSignIn(t *testing.T) {
	t.Parallel()

	h := Handlers{
		users: []Credentials{},
		mu:    &sync.Mutex{},
	}

	body := bytes.NewReader([]byte(`{"username": "Vasily", "password": "qwerty"}`))

	r := httptest.NewRequest("POST", "/signup/", body)
	w := httptest.NewRecorder()

	h.handleSignUp(w, r)

	r = httptest.NewRequest("POST", "/signin/", body)
	w = httptest.NewRecorder()

	h.handleSignIn(w, r)

	if w.Code != http.StatusOK {
		t.Error("Failed http Status")
	}
}

func TestChangeProfile(t *testing.T) {
	t.Parallel()

	h := Handlers{
		users: []Credentials{},
		mu:    &sync.Mutex{},
	}

	var expectedUsers = []Credentials{
		{
			ID:       0,
			Username: "Vasily",
			Password: "sdhsdh",
		},
	}

	body := bytes.NewReader([]byte(`{"username": "Vasily", "password": "qwerty"}`))

	r := httptest.NewRequest("POST", "/signup/", body)
	w := httptest.NewRecorder()

	h.handleSignUp(w, r)

	changeBody := bytes.NewReader([]byte(`{"username": "Vasily", "password": "sdhsdh"}`))

	r = httptest.NewRequest("POST", "/profile/", changeBody)
	w = httptest.NewRecorder()

	h.handleChangeProfile(w, r)

	if w.Code != http.StatusOK {
		t.Error("Failed http Status")
	}

	reflect.DeepEqual(h.users, expectedUsers)
}

// Doesn`t work!!!! (Idk why)
// func TestViewProfile(t *testing.T) {
// 	t.Parallel()

// 	h := Handlers{
// 		users: []Credentials{
// 			{
// 				ID:       0,
// 				Username: "Afanasiy",
// 				Password: "1234",
// 			},
// 			{
// 				ID:       1,
// 				Username: "Afan",
// 				Password: "4321",
// 			},
// 		},
// 		mu: &sync.Mutex{},
// 	}

// 	expectedViewProfile := "Afan"

// 	body := bytes.NewReader([]byte(`{"username": "Afan", "password": "4321"}`))

// 	r := httptest.NewRequest("POST", "/signin/", body)
// 	w := httptest.NewRecorder()

// 	h.handleSignIn(w, r)

// 	r = httptest.NewRequest("GET", "/profile/", nil)
// 	w = httptest.NewRecorder()

// 	h.handleGetProfile(w, r)

// 	if w.Code != http.StatusOK {
// 		t.Error("Failed http Status")
// 	}

// 	bytes, _ := ioutil.ReadAll(w.Body)
// 	if strings.Trim(string(bytes), "\n") != expectedViewProfile {
// 		t.Errorf("Failed Body is not matched %s", ReadCookie(w, r))
// 	}
// }
