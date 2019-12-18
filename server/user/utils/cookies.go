package utils

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/auth/session"
	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/model"
	"github.com/gorilla/securecookie"
	"github.com/labstack/echo"
)

var (
	SessManager session.AuthCheckerClient
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)

func SetSession(ctx echo.Context, userData *model.User) (*http.Cookie, error) {
	sessID, err := SessManager.Create(
		context.Background(),
		&session.Session{
			Username: userData.Username,
			Avatar:   userData.Avatar,
		})
	if err != nil {
		return nil, err
	}
	value := map[string]string{
		"session_id": sessID.ID,
		"user_id":    strconv.Itoa(int(userData.ID)),
	}

	if encoded, err := cookieHandler.Encode("session_token", value); err == nil {
		expiration := time.Now().Add(24 * time.Hour)
		cookie := &http.Cookie{
			Name:    "session_token",
			Value:   encoded,
			Path:    "/",
			Expires: expiration,
		}
		ctx.SetCookie(cookie)
		return cookie, nil
	}
	return nil, nil
}

func ClearSession(ctx echo.Context) error {
	_, err := SessManager.Delete(
		context.Background(),
		ReadSessionID(ctx))
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}
	ctx.SetCookie(cookie)
	return nil
}

func СheckSession(ctx echo.Context) (*session.Session, error) {
	cookieSessionID := ReadSessionID(ctx)
	log.Println(cookieSessionID)
	if cookieSessionID == nil {
		return nil, nil
	}

	sess, err := SessManager.Check(
		context.Background(),
		cookieSessionID)
	if err != nil {
		log.Println(err)
	}
	return sess, nil
}

func ReadSessionID(ctx echo.Context) *session.SessionID {
	if cookie, err := ctx.Request().Cookie("session_token"); err == nil {
		log.Println(cookie.Value)
		value := make(map[string]string)
		if err = cookieHandler.Decode("session_token", cookie.Value, &value); err == nil {
			return &session.SessionID{ID: value["session_id"]}
		}
	}
	return nil
}

func ReadSessionIDAndUserID(ctx echo.Context) []string {
	if cookie, err := ctx.Request().Cookie("session_token"); err == nil {
		value := make(map[string]string)
		if err = cookieHandler.Decode("session_token", cookie.Value, &value); err == nil {
			var result = []string{}
			result = append(result, value["session_id"])
			result = append(result, value["user_id"])
			return result
		}
	}
	return nil
}

func ReadSessionIDAndUserIDJWT(cookie *http.Cookie) []string {
	value := make(map[string]string)
	if err := cookieHandler.Decode("session_token", cookie.Value, &value); err == nil {
		var result = []string{}
		result = append(result, value["session_id"])
		result = append(result, value["user_id"])
		return result
	}
	return nil
}
