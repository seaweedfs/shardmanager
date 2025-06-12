.PHONY: test

SOURCE_DIR = .
debug ?= 0

PROTO_SRC=shardmanager.proto
PROTO_OUT=./shardmanagerpb

all: proto build

proto:
	protoc --go_out=$(PROTO_OUT) --go-grpc_out=$(PROTO_OUT) \
	  --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative \
	  $(PROTO_SRC)

build:
	go build ./...

test:
	go test ./...

clean:
	rm -rf $(PROTO_OUT)/*.pb.go
