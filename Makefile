PROTO_DIR=pkg/proto
OUT_DIR=pkg/proto/storagepb

PROTOC=protoc
GOOGLE_PROTOBUF=$(shell go env GOPATH)/pkg/mod
PROTO_FILES=$(PROTO_DIR)/*.proto

generate:
	$(PROTOC) \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

clean:
	rm -rf $(OUT_DIR)/*.go