package main

import (
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/labstack/echo"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)

// SetCookie - set cookie with gluing together secret word and username login with ecryption
func SetCookie(ctx echo.Context, userName string) {
	value := map[string]string{
		"username": userName,
	}

	if encoded, err := cookieHandler.Encode(sessionSecret, value); err == nil {
		expiration := time.Now().Add(24 * time.Hour)
		cookie := http.Cookie{
			Name:    "session_token",
			Value:   encoded,
			Path:    "/",
			Expires: expiration,
		}

		http.SetCookie(ctx.Response(), &cookie)
	}
}

// ClearCookie - delete cookie by set time in past date
func ClearCookie(ctx echo.Context) {
	cookie := http.Cookie{
		Name:    "session_token",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(ctx.Response(), &cookie)
}

// ReadCookieUsername - return login by decoding cookie with secret word or empty string if not exist
func ReadCookieUsername(ctx echo.Context) string {
	if cookie, err := ctx.Request().Cookie(sessionSecret); err == nil {
		value := make(map[string]string)
		if err = cookieHandler.Decode(sessionSecret, cookie.Value, &value); err == nil {
			return value["username"]
		}
	}
	return ""
}

// ReadCookieAvatar - return avatar by decoding cookie with secret word and pulling it from DataBase or return empty string
func (h *Handlers) ReadCookieAvatar(ctx echo.Context) string {
	if cookie, err := ctx.Request().Cookie(sessionSecret); err == nil {
		value := make(map[string]string)
		if err = cookieHandler.Decode(sessionSecret, cookie.Value, &value); err == nil {
			userRecord, err := h.Users.SelectDataByLogin(value["username"])
			if err != nil {
				return ""
			}

			return userRecord.Avatar
		}
	}
	return ""
}
