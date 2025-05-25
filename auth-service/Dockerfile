FROM golang:1.24 as builder



WORKDIR /app

COPY . .





RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /auth_service cmd/main.go


FROM alpine:3.21


WORKDIR /app

COPY  --from=builder /auth_service .

 
COPY --from=builder /app/db  ./db 

COPY --from=builder /app/certs  ./certs 


EXPOSE 8090

CMD [ "./auth_service" ]