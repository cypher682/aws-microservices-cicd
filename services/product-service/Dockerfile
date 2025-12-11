FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

RUN addgroup -g 1001 -S appuser && adduser -S appuser -u 1001 -G appuser
USER appuser

EXPOSE 8080

CMD ["./main"]
