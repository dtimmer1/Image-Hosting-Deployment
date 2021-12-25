package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"
)

type EventType int

const (
	UndefinedEvent EventType = iota
	ImageRequestEvent
	ImageUploadEvent
)

const MAX_BUFFER_SIZE = 0

type Event struct {
	Type    EventType
	Request string
	Reply   chan string
}

type QueryHandler struct {
	eventQueue    chan Event
	quit          chan struct{}
	imageDatabase map[string]string
}

func NewQueryHandler() *QueryHandler {
	return &QueryHandler{make(chan Event, MAX_BUFFER_SIZE), make(chan struct{}), make(map[string]string)}
}

func (qh *QueryHandler) Launch() {
	for {
		select {
		case currEvent := <-qh.eventQueue:
			switch currEvent.Type {
			case ImageRequestEvent:
				log.Printf("requested image is %s\n", currEvent.Request)
				currEvent.Reply <- qh.upload(currEvent.Request)

			case ImageUploadEvent:
				hasher := sha1.New()
				hasher.Write([]byte(currEvent.Request))
				encoding := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
				link := strings.ReplaceAll(encoding, "/", "")[:10]

				qh.imageDatabase[link] = currEvent.Request
				currEvent.Reply <- link
			default:
				log.Fatal("Undefined event, this should not happen")
			}

		case <-qh.quit:
			return
		}
	}
}

func (qh *QueryHandler) upload(s string) string {
	img, ok := qh.imageDatabase[s]

	if !ok {
		return ""
	}

	idx := strings.Index(img, ",")

	decoded, err := base64.StdEncoding.DecodeString(img[idx+1:])

	if err != nil {
		log.Println(err)
		return ""
	}

	image, format, err := image.Decode(bytes.NewReader(decoded))

	if err != nil {
		log.Println(err)
		return ""
	}

	filename := "./static/img/" + s + "." + format

	_, err = os.Open(filename) //check if file already exists

	if err == nil { //file found
		return filename
	}
	f, err := os.Create(filename)

	if err != nil {
		log.Println(err)
		return ""
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
		return ""
	}

	f.Close()

	return filename
}

func (qh *QueryHandler) SendEvent(e Event) {
	qh.eventQueue <- e
}

func (qh *QueryHandler) Quit() {
	qh.quit <- struct{}{}
}
