FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build -o comments ./cmd/server

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/comments ./comments
EXPOSE 8082
ENTRYPOINT ["./comments"]