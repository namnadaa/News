FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./
RUN go build -o news ./cmd/server

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/news .
RUN chmod +x ./news
EXPOSE 8080
ENTRYPOINT ["./news"]

# FROM --platform=linux/arm64 golang:1.24.3-alpine AS builder

# WORKDIR /app
# COPY go.mod ./ 
# COPY go.sum ./
# RUN go mod download
# COPY . ./
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o news ./cmd/server

# FROM --platform=linux/arm64 alpine:latest
# WORKDIR /root/
# COPY --from=builder /app/news .
# RUN chmod +x ./news
# EXPOSE 8080
# ENTRYPOINT ["./news"]