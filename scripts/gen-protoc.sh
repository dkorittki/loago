#!/usr/bin/env bash

protoc -I ./api/v1/ -I ${GOPATH}/src \
    ./api/v1/worker.proto --go_out=plugins=grpc:pkg/api/v1/ --govalidators_out=pkg/api/v1/