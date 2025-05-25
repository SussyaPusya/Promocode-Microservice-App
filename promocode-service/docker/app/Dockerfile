FROM golang:1.24 as builder

WORKDIR /app

# Gitlab auth setup into ... accessToken
RUN echo "machine gitlab.com\nlogin oauth2\npassword ..." > ~/.netrc
RUN chmod 600 ~/.netrc

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/app/main.go

FROM debian:bullseye-slim
WORKDIR /root/

RUN apt-get update && \
    apt-get install -y curl ca-certificates postgresql-client dos2unix && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin && \
    chmod +x /usr/local/bin/migrate && \
     apt-get remove -y curl && apt-get clean

COPY --from=builder /app/main .
COPY --from=builder /app/.env .
COPY --from=builder /app/migrations ./migrations



EXPOSE 50060

CMD [ "./main" ]
