FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o MerchServiceAvito ./cmd/server/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/MerchServiceAvito .
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations  
COPY .env .  

RUN apk add --no-cache bash curl
RUN curl -o wait-for-it.sh https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh
RUN chmod +x wait-for-it.sh

EXPOSE 8080
CMD ["./wait-for-it.sh", "db:5432", "--", "./MerchServiceAvito"]