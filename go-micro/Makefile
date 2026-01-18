ADDR = localhost:8080
PROTO_PATH = ./proto/store/store.proto
GRPCURL = $(shell which grpcurl)

.PHONY: proto-store proto-file-transaction-logger get put del

proto-store: 
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	./proto/store/store.proto

proto-file-transaction-logger:
	protoc --go_out=. --go_opt=paths=source_relative \
	./proto/transactionLogger/transactionLogger.proto

## put: Store a key-value pair. Usage: make put KEY=foo VAL=bar
put:
	@$(GRPCURL) -plaintext -d '{"key": "$(KEY)", "value": "$(VAL)"}' $(ADDR) store.StoreService/PutHandler

## get: Retrieve a value. Usage: make get KEY=foo
get:
	@$(GRPCURL) -plaintext -d '{"key": "$(KEY)"}' $(ADDR) store.StoreService/GetHandler

## del: Delete a key. Usage: make del KEY=foo
del:
	@$(GRPCURL) -plaintext -d '{"key": "$(KEY)"}' $(ADDR) store.StoreService/DelHandler