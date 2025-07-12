package handlers

import (
	"github.com/labstack/echo/v4"
	"goodoo/templates"
)

func IndexHandler(c echo.Context) error {
	title := "Hello from Goodoo!"
	message := "This is a sample Echo + Templ application. The templating engine provides type-safe HTML generation with excellent performance."
	
	component := templates.IndexPage(title, message)
	return component.Render(c.Request().Context(), c.Response().Writer)
}