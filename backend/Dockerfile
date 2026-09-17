FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o server ./cmd/server
FROM golang:1.23-alpine
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
