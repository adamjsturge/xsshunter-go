FROM golang:1.23-alpine3.19 as builder

RUN apk update && apk add --no-cache gcc musl-dev ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG GIT_TAG
ARG GIT_COMMIT
ARG GIT_BRANCH

RUN BUILD_DATE=$(date +'%Y-%m-%dT%H:%M:%S%z') && \
    go build -ldflags "-X 'main.version=${GIT_TAG}' -X 'main.gitCommit=${GIT_COMMIT}' -X 'main.gitBranch=${GIT_BRANCH}' -X 'main.buildDate=${BUILD_DATE}'" -o main

FROM alpine:3.20.2 as prod
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY probe.js .
COPY src ./src

EXPOSE 1449

CMD ["./main"]

FROM golang:1.23-alpine3.19 as dev

RUN apk update && apk add --no-cache gcc musl-dev ca-certificates

RUN go install github.com/air-verse/air@latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

CMD ["air", "-c", ".air.toml"]