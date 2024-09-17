FROM golang:1.23.1 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go install github.com/moderato-app/live-pprof@v1
RUN go mod download

COPY . .

RUN go build -o main .

FROM golang:1.23.1

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8081

CMD ["./main"]
