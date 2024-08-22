# Build stage 
FROM golang:1.22.6 AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o bin .

# Final stage
FROM busybox:latest

COPY --from=builder /app/bin /chatroom/

EXPOSE 8080

CMD ["/chatroom/bin"]