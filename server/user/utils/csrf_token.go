package utils

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo"

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
func (tk *JwtToken) Create(s []string, tokenExpTime int64) (string, error) {
	userID, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		return "", nil
	}
	data := JwtCsrfClaims{
		SessionID: s[0],
		UserID:    userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: tokenExpTime,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	return token.SignedString(tk.Secret)
}

// Check - checks signing token
func (tk *JwtToken) Check(s []string, inputToken string) (bool, error) {
	payload := &JwtCsrfClaims{}
	_, err := jwt.ParseWithClaims(inputToken, payload, tk.parseSecretGetter)
	if err != nil {
		return false, fmt.Errorf("cant parse jwt token: %v", err)
	}
	if payload.Valid() != nil {
		return false, fmt.Errorf("invalid jwt token: %v", err)
	}
	userID, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		return false, nil
	}
	return payload.SessionID == s[0] && payload.UserID == userID, nil
}

// parseSecretGetter - parse single secret token
func (tk *JwtToken) parseSecretGetter(token *jwt.Token) (interface{}, error) {
	method, ok := token.Method.(*jwt.SigningMethodHMAC)
	if !ok || method.Alg() != "HS256" {
		return nil, fmt.Errorf("bad sign method")
	}
	return tk.Secret, nil
}

// SetToken - set and sign token and return token and error
func SetToken(ctx echo.Context, cookieSes *http.Cookie) error {
	session := ReadSessionIDAndUserIDJWT(cookieSes)
	log.Println(session)

	token, err := NewJwtToken(Secret)
	log.Println(token)
	if err != nil {
		return err
	}

	expiresAt := int64(60 * 360)

	t, err := token.Create(session, expiresAt)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:    "Csrf-Token",
		Value:   t,
		Expires: time.Now().AddDate(0, 0, 1),
	}

	ctx.SetCookie(cookie)

	return nil
}
