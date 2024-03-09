FROM golang:1.22-alpine

RUN apk update && apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG GIT_TAG
ARG GIT_COMMIT
ARG GIT_BRANCH

RUN BUILD_DATE=$(date +'%Y-%m-%dT%H:%M:%S%z') && \
    go build -ldflags "-X 'main.version=${GIT_TAG}' -X 'main.gitCommit=${GIT_COMMIT}' -X 'main.gitBranch=${GIT_BRANCH}' -X 'main.buildDate=${BUILD_DATE}'" -o main

EXPOSE 1449

CMD ["./main"]