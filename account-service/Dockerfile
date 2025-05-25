FROM golang:1.23.7-alpine3.21 as builder



WORKDIR /app

COPY . .





RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /account_service cmd/main.go


FROM alpine:3.21


WORKDIR /app

COPY  --from=builder /account_service .

 
COPY --from=builder /app/db  ./db 

COPY --from=builder /app/certs  ./certs 


EXPOSE 8090

EXPOSE 50051


CMD [ "./account_service" ]