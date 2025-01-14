FROM golang:1.22.0-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o kubectl-yadt

FROM alpine:3.19

COPY --from=builder /app/kubectl-yadt /usr/local/bin/kubectl-yadt

ENTRYPOINT ["kubectl-yadt"]