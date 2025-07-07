FROM golang:1.24 as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o netchecknova ./cmd/netchecknova

FROM alpine:3.20

WORKDIR /app
RUN apk add --no-cache ca-certificates
RUN chgrp -R 0 /app && \
    chmod -R g+rwX /app
COPY --from=builder /app/netchecknova .
CMD ["./netchecknova"]
