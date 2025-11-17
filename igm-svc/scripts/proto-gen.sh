#!/bin/bash
# filepath: scripts/proto-gen.sh
set -e
PROTO_DIR="api/proto/igm/v1"
PROTO_FILE="$PROTO_DIR/issue.proto"

echo "generating prtobuf code for $PROTO_FILE"

protoc \
    --go_out=. \
    --go_opt=paths=source_relative \
    --go-grpc_out=. \
    --go-grpc_opt=paths=source_relative \
    $PROTO_FILE

echo "generated"