
FROM golang:1.23-alpine
WORKDIR /app
RUN apk add --no-cache git ca-certificates
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server
RUN chmod +x main
EXPOSE 8080
CMD ["./main"]