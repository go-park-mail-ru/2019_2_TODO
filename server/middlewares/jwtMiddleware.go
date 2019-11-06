package middlewares

import (
	"net/http"
	"server/user/utils"
	"strings"

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
		expiresAt := int64(60 * 360)
		token, err := tokenHandler.Create(session, expiresAt)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, "Token Create")
		}

		c, err := ctx.Request().Cookie("Csrf-Token")
		if err != nil {
			return err
		}
		tokenFront := c.Value

		if tokenFront[:strings.IndexByte(tokenFront, '.')] != token[:strings.IndexByte(token, '.')] {
			return ctx.JSON(http.StatusBadRequest, "Wrong Token")
		}

		return next(ctx)
	}
}
