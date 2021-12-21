package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var imgs map[string]string

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func writeImageAndServe(c echo.Context) error {
	img, ok := imgs[c.Param("link")]

	if !ok {
		return c.Render(http.StatusNotFound, "notfound.html", map[string]interface{}{})
	}

	idx := strings.Index(img, ",")

	decoded, err := base64.StdEncoding.DecodeString(img[idx+1:])

	if err != nil {
		log.Println(err)
		return c.Render(http.StatusNotFound, "notfound.html", map[string]interface{}{})
	}

	image, format, err := image.Decode(bytes.NewReader(decoded))

	if err != nil {
		log.Println(err)
		return c.Render(http.StatusNotFound, "notfound.html", map[string]interface{}{})
	}

	filename := "./static/img/" + c.Param("link") + "." + format

	_, err = os.Open(filename) //check if file already exists

	if err == nil { //file found
		return c.File(filename)
	}

	f, err := os.Create(filename)

	if err != nil {
		log.Println(err)
		log.Println("ducky")
		return c.Render(http.StatusNotFound, "notfound.html", map[string]interface{}{})
	}

	switch format {
	case "jpeg":
		err = jpeg.Encode(f, image, nil)
	case "jpg":
		err = jpeg.Encode(f, image, nil)
	case "png":
		err = png.Encode(f, image)
	case "gif":
		err = gif.Encode(f, image, nil)
	}

	if err != nil {
		log.Println(err)
		return c.Render(http.StatusNotFound, "notfound.html", map[string]interface{}{})
	}

	f.Close()

	return c.File(filename)
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	funcMap := template.FuncMap{
		"safe": func(s string) template.URL {
			return template.URL(s)
		},
	}

	t := &TemplateRenderer{
		templates: template.Must(template.New("").Funcs(funcMap).ParseGlob("html/*.html")),
	}

	e.Renderer = t

	e.Static("/", "static")

	imgs = make(map[string]string)

	e.GET("/hello/:name", func(c echo.Context) error {
		name := c.Param("name")
		return c.Render(http.StatusOK, "hello.html", map[string]interface{}{"hello": name})
	})

	e.GET("/", func(c echo.Context) error {
		return c.File("/html/img.html")
	})

	e.POST("/ping", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct {
			Status string `json:"status"`
		}{Status: "ok"})
	})

	e.POST("/send", func(c echo.Context) error {
		body := c.Request().Body
		var data struct {
			Image string `json:"image"`
		}
		decoder := json.NewDecoder(body)
		decoder.Decode(&data)
		hasher := sha1.New()
		hasher.Write([]byte(data.Image))
		encoding := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
		link := strings.ReplaceAll(encoding, "/", "")[:10]

		imgs[link] = data.Image

		return c.String(http.StatusOK, link)
	})

	e.GET("/:link", writeImageAndServe)

	e.Logger.Fatal(e.Start(":8080"))
}
