# syntax=docker/dockerfile:1

FROM golang:1.16-alpine AS build

WORKDIR /build

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
COPY /html/ /html/
COPY /static/ /static/

RUN go build -o /test-server

FROM alpine:latest
COPY  --from=build /test-server /test-server
COPY --from=build /html/ /html/
COPY --from=build /static/ /static/

EXPOSE 8080
CMD ["/test-server"]