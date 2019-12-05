package middlewares

import (
	"net/http"
	"strings"

	"github.com/go-park-mail-ru/2019_2_TODO/tree/devRK/server/user/utils"

	"github.com/labstack/echo"
)

// JWTMiddlewareCustom - checks jwt token
func JWTMiddlewareCustom(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		tokenHandler := utils.JwtToken{}
		session := utils.ReadSessionIDAndUserID(ctx)

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
