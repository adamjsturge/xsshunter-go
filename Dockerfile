FROM golang:1.22
# AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

# FROM golang:1.22-alpine
# WORKDIR /app
# COPY --from=builder /app/main ./main

EXPOSE 8080

CMD ["./main"]