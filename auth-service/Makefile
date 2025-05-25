
PROTO_PATH := api/protos
OUTPUT_PATH := .
BINARY_NAME := myapp

# Цель по умолчанию
all: build

# Генерация кода из .proto файлов
generateAccountProto:
	protoc --go_out=$(OUTPUT_PATH) --go-grpc_out=$(OUTPUT_PATH)  --openapiv2_out $(OUTPUT_PATH)  $(PROTO_PATH)/account.proto 


generatePromoProto:
	protoc --go_out=$(OUTPUT_PATH) --go-grpc_out=$(OUTPUT_PATH)  --openapiv2_out $(OUTPUT_PATH)  $(PROTO_PATH)/promo.proto 

# Сборка бинарника
build: generate
	go build -o $(BINARY_NAME) cmd/main.go

# Очистка сгенерированных файлов и бинарника
clean:
	rm -rf $(OUTPUT_PATH) $(BINARY_NAME)

# Установка необходимых инструментов
install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

.PHONY: all generateAccountProto generatePromoProto build clean install-tools