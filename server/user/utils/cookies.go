package utils

import (
	"log"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/model"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
)

// SessionsStore - hold all cookies
var SessionsStore = sessions.NewCookieStore(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)

// SetCookie - set cookie with necessary data
func SetCookie(ctx echo.Context, userInfo model.User) error {
	session, err := SessionsStore.Get(ctx.Request(), "session_token")
	if err != nil {
		return err
	}

	session.Values["id"] = userInfo.ID
	session.Values["username"] = userInfo.Username
	session.Values["avatar"] = userInfo.Avatar

	err = session.Save(ctx.Request(), ctx.Response())
	if err != nil {
		return err
	}
	return nil
}

// ClearCookie - delete cookie by set time in past date
func ClearCookie(ctx echo.Context) error {
	session, err := SessionsStore.Get(ctx.Request(), "session_token")
	if err != nil {
		return err
	}

	session.Values["ID"] = nil
	session.Values["username"] = ""
	session.Values["avatar"] = ""
	session.Options.MaxAge = -1

	err = session.Save(ctx.Request(), ctx.Response())
	if err != nil {
		return err
	}
	return nil
}

// ReadCookieID - return ID from cookie
func ReadCookieID(ctx echo.Context) int64 {
	session, err := SessionsStore.Get(ctx.Request(), "session_token")
	if err != nil {
		return -1
	}
	if session.Values["ID"] == nil {
		return -1
	}
	return session.Values["ID"].(int64)
}

// ReadCookieUsername - return username from cookie
func ReadCookieUsername(ctx echo.Context) string {
	session, err := SessionsStore.Get(ctx.Request(), "session_token")
	log.Println(err)
	if err != nil {
		return ""
	}
	if session.Values["username"] == nil {
		return ""
	}
	return session.Values["username"].(string)
}

// ReadCookieAvatar - return avatar from cookie
func ReadCookieAvatar(ctx echo.Context) string {
	session, err := SessionsStore.Get(ctx.Request(), "session_token")
	if err != nil {
		return ""
	}
	if session.Values["avatar"] == nil {
		return ""
	}
	return session.Values["avatar"].(string)
}
