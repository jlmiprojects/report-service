package utils

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

const secret = "Self@Track@123456_JLMIProjects_2007"

func CreateJWTConfig() echo.MiddlewareFunc {

	return echojwt.JWT([]byte(secret))

}
