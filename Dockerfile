FROM golang:1.21 as builder
WORKDIR /app
COPY go.mod go.sum ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o rate-limiter .

FROM scratch
WORKDIR /app
COPY --from=builder /app/rate-limiter .
CMD ["./rate-limiter"]