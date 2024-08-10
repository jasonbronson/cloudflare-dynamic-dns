FROM golang:1.22.5-bullseye

RUN mkdir -p /app
WORKDIR /app

ADD . /app

ENV GO111MODULE=on

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ./main /app/main.go

CMD ["/app/main"]
