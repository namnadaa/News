FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build -o news ./cmd/server

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/news ./news
COPY config.json ./config.json
COPY webapp ./webapp
EXPOSE 8080
ENTRYPOINT ["./news"]
