FROM golang:1.25.1-alpine3.18 AS builder

WORKDIR /app
COPY . .

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn,direct \

RUN go build -o main .

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /app/main .
CMD ["./main"]