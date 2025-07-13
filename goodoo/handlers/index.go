package handlers

import (
	"github.com/labstack/echo/v4"
)

func IndexHandler(c echo.Context) error {
	return c.File("templates/home.html")
}

func LoginPageHandler(c echo.Context) error {
	return c.File("templates/login.html")
}