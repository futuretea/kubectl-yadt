FROM golang:1.22.0-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o wtfk8s

FROM alpine:3.19

COPY --from=builder /app/wtfk8s /usr/local/bin/wtfk8s

ENTRYPOINT ["wtfk8s"] 