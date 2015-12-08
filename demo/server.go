package main

import (
	"net/http"

	"github.com/venjiang/echo"
)

func index(c *echo.Context) error {
	return c.Render(http.StatusOK, "index.html", "hello")
}

func view(c *echo.Context) error {
	c.ViewData["msg"] = "Wecome to echo ViewData"
	return c.View(http.StatusOK, "shared/view.html")
}

func main() {
	e := echo.New()
	e.Debug()
	e.SetRenderer(echo.HtmlRenderer(echo.Options{Layout: "layout/layout.html"}))

	e.Get("/", index)
	e.Get("/view", view)

	e.Run(":1326")
}
