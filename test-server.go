package main

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	t := &TemplateRenderer{
		templates: template.Must(template.New("").ParseGlob("html/*.html")),
	}

	e.Renderer = t

	e.Static("/", "static")

	qh := NewQueryHandler()

	go qh.Launch()

	defer qh.Quit()

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "img.html", map[string]interface{}{})
	})

	e.POST("/send", func(c echo.Context) error {
		body := c.Request().Body
		var data struct {
			Image string `json:"image"`
		}
		decoder := json.NewDecoder(body)
		decoder.Decode(&data)

		eventRequest := make(chan string)

		qh.SendEvent(Event{ImageUploadEvent, data.Image, eventRequest})

		var output string

		select {
		case output = <-eventRequest:
			return c.String(http.StatusOK, output)
		case <-time.After(time.Duration(time.Second * 5)):
			return c.String(http.StatusInternalServerError, "server timed out")
		}
	})

	e.GET("/img/:link", func(c echo.Context) error {
		log.Println(c.Param("link"))
		eventRequest := make(chan string)

		qh.SendEvent(Event{ImageRequestEvent, c.Param("link"), eventRequest})

		select {
		case output := <-eventRequest:
			if output != "" {
				return c.File(output)
			} else {
				return c.File("html/notfound.html")
			}
		case <-time.After(time.Duration(time.Second * 5)):
			return c.File("html/notfound.html")
		}
	})

	e.Logger.Fatal(e.Start(":8080"))
}
