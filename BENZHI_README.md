# go-property-rental Evaluation Guide

## Project

- Repository: `zhangkui/go-property-rental`
- Branch: `test_model_fix26`
- Go toolchain: `golang:1.22`

## Standard Commands

```bash
cd /app && GOTOOLCHAIN=local go build ./...
cd /app && GOTOOLCHAIN=local go test ./...
cd /app && GOTOOLCHAIN=local go run ./cmd/api
```

## Docker Build

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh go-property-rental-bug026-candidate-amd64 linux/amd64
docker run --rm -it --platform linux/amd64 go-property-rental-bug026-candidate-amd64 bash
```

## Public Verification

```bash
go test ./scripts/verify -count=1 -run '^TestBug026_BusinessRegression$'
```

This branch contains production code, runtime files, and the public verification test only.
