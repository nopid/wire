FROM golang:alpine AS builder
WORKDIR /src
COPY wire.go .
COPY go.mod .
RUN go mod tidy
RUN go mod download
RUN GOOS=linux go build .

FROM alpine:latest
RUN apk --no-cache add tcpdump
WORKDIR /app
COPY --from=builder /src/wire .
CMD ["./wire"]  
