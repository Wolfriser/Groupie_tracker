FROM golang:1.20

COPY . /

WORKDIR /

RUN go build main.go

CMD ["./main"]

EXPOSE 8080