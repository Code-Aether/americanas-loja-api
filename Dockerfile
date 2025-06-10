FROM golang:1.23-alpine AS builder

ARG APP_NAME=api
WORKDIR /app

RUN apk add --no-cache make git build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

RUN [ -f ./bin/${APP_NAME} ] || (echo "Binário não foi criado" && exit 1)
RUN chmod +x ./bin/api

FROM alpine:latest

ARG APP_NAME=api
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/${APP_NAME} .
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./api"]