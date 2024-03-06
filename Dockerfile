FROM golang:1.22
# AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# RUN go build -o main .

RUN go build -ldflags "-X 'main.version=$(git describe --tags --abbrev=0)' -X 'main.gitCommit=$(git rev-parse HEAD)' -X 'main.gitBranch=$(git rev-parse --abbrev-ref HEAD)' -X 'main.buildDate=$(date +'%Y-%m-%dT%H:%M:%S%z')'" -o main

# FROM golang:1.22-alpine
# WORKDIR /app
# COPY --from=builder /app/main ./main


EXPOSE 1449

CMD ["./main"]