FROM golang:alpine3.19 AS builder

ENV CGO_ENABLED 0

ENV GOOS linux

RUN apk update --no-cache

WORKDIR /builder

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

RUN go build main.go

FROM alpine:3.19.1 AS runner

RUN apk update --no-cache && apk add --no-cache ca-certificates && apk add --no-cache --upgrade bash

WORKDIR /build

COPY . .

COPY --from=builder /builder/main /build

CMD ["./main"]