package main

import (
	"fmt"
	"time"

	"github.com/gorilla/sessions"

	jwt "github.com/dgrijalva/jwt-go"
)

// Key - struct for sliding secret keys
type Key struct {
	Exp    int64
	Secret string
}

// JwtToken - JSON object secret
type JwtToken struct {
	Secret []byte
}

// NewJwtToken - create JwtToken
func NewJwtToken(secret string) (*JwtToken, error) {
	return &JwtToken{Secret: []byte(secret)}, nil
}

// JwtCsrfClaims - records of csrf token
type JwtCsrfClaims struct {
	SessionID string `json:"sid"`
	UserID    int64  `json:"uid"`
	jwt.StandardClaims
}

// Create - create and sign JwtToken
func (tk *JwtToken) Create(s *sessions.Session, tokenExpTime int64) (string, error) {
	data := JwtCsrfClaims{
		SessionID: s.ID,
		UserID:    s.Values["id"].(int64),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: tokenExpTime,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	return token.SignedString(tk.Secret)
}

// Check - checks signing token
func (tk *JwtToken) Check(s *sessions.Session, inputToken string) (bool, error) {
	payload := &JwtCsrfClaims{}
	_, err := jwt.ParseWithClaims(inputToken, payload, tk.parseSecretGetterMultiKeys)
	if err != nil {
		return false, fmt.Errorf("cant parse jwt token: %v", err)
	}
	if payload.Valid() != nil {
		return false, fmt.Errorf("invalid jwt token: %v", err)
	}
	return payload.SessionID == s.ID && payload.UserID == s.Values["ID"], nil
}

// parseSecretGetter - parse single secret token
func (tk *JwtToken) parseSecretGetter(token *jwt.Token) (interface{}, error) {
	method, ok := token.Method.(*jwt.SigningMethodHMAC)
	if !ok || method.Alg() != "HS256" {
		return nil, fmt.Errorf("bad sign method")
	}
	return tk.Secret, nil
}

// parseSecretGetterMultiKeys - parse multi secret tokens
func (tk *JwtToken) parseSecretGetterMultiKeys(token *jwt.Token) (interface{}, error) {
	method, ok := token.Method.(*jwt.SigningMethodHMAC)
	if !ok || method.Alg() != "HS256" {
		return nil, fmt.Errorf("bad sign method")
	}
	keys := []*Key{
		&Key{Exp: 10, Secret: "abs"},
		&Key{Exp: 20, Secret: "hsdh"},
		&Key{Exp: 30, Secret: "aosgja"},
	}
	payload, ok := token.Claims.(*JwtCsrfClaims)
	if !ok {
		return nil, fmt.Errorf("Claims error")
	}
	secret := ""
	for _, key := range keys {
		if key.Exp > payload.ExpiresAt {
			secret = key.Secret
			break
		}
	}
	if secret == "" {
		return nil, fmt.Errorf("no secret found")
	}
	return secret, nil
}
