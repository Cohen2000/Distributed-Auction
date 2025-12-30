FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /coordinator ./cmd/coordinator
RUN CGO_ENABLED=0 go build -o /node ./cmd/node

FROM alpine:3.19
COPY --from=builder /coordinator /coordinator
COPY --from=builder /node /node
