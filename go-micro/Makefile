proto-store: 
	protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       ./proto/store/store.proto

proto-file-transaction-logger:
       protoc --go_out=. --go_opt=paths=source_relative \
       ./proto/transactionLogger/transactionLogger.proto

.PHONY: proto-store, proto-file-transaction-logger