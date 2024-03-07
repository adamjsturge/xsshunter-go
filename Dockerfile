FROM golang:1.22

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG GIT_TAG
ARG GIT_COMMIT
ARG GIT_BRANCH
ARG BUILD_DATE

RUN go build -ldflags "-X 'main.version=${GIT_TAG}' -X 'main.gitCommit=${GIT_COMMIT}' -X 'main.gitBranch=${GIT_BRANCH}' -X 'main.buildDate=${BUILD_DATE}'" -o main

EXPOSE 1449

CMD ["./main"]