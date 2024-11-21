#!/bin/bash

# Create necessary directories
mkdir -p internal/payload/{auth,base,error,heartbeat,term,info}

# First, create individual proto files in a protos directory
mkdir -p protos

# Generate the protobuf code
protoc \
    --proto_path=protos \
    --go_out=. \
    --go_opt=module=willofdaedalus/superluminal \
    protos/*.proto

# Optional: Clean up temporary protos directory
# rm -rf protos
