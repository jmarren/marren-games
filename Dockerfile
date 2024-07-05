#
#
#
FROM golang:1.22-alpine


WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify
RUN apk update \
  && apk add --no-cache sqlite \ 
  && apk --no-cache --update add build-base

COPY . .
RUN go env -w GOARCH=arm64
RUN go build -o ./tmp/main ./cmd/server

EXPOSE 8080

CMD ["./tmp/main"]
