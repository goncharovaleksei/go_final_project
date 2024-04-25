FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
COPY web ./web

EXPOSE 7540

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /my_app

CMD ["/my_app"]

ENV TODO_PORT 7540
ENV TODO_PASSWORD testpas
ENV TODO_DBFILE ./scheduler.db