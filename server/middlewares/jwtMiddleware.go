package middlewares

import (
	"net/http"
	"server/user/utils"

	"github.com/labstack/echo"
)

// JWTMiddlewareCustom - checks jwt token
func JWTMiddlewareCustom(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		tokenHandler := utils.JwtToken{}
		session, err := utils.SessionsStore.Get(ctx.Request(), "session_token")
		if err != nil {
			return err
		}

		token := ctx.Request().Header.Get("csrf-token")

		ok, err := tokenHandler.Check(session, token)
		if !ok {
			return ctx.JSON(http.StatusBadRequest, "wrong csrf-token")
		}
		if err != nil {
			return err
		}

		return next(ctx)
	}
}
