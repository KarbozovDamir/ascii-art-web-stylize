FROM golang:1.16.3

LABEL project="ASCII-ART-WEB"

WORKDIR /web

COPY . .

RUN go build -o main

EXPOSE 8080

CMD ["/web/main"]