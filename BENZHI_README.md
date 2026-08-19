# go-property-rental 评测说明

## 项目说明
- 项目：`zhangkui/go-property-rental`
- 用途：长租房源与租约管理平台。
- Go 工具链：`golang:1.22`
- 前端工具链：Node.js 22、Vue 3、TypeScript、Vite。

## 标准构建、运行和测试命令
```bash
cd '/app' && GOTOOLCHAIN=local go build ./...
cd '/app' && GOTOOLCHAIN=local go test ./...
cd '/app' && GOTOOLCHAIN=local go run ./cmd/api
```

## Docker 构建和进入容器
```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh go-property-rental-bug3-candidate-amd64 linux/amd64
docker run --rm -it --platform linux/amd64 go-property-rental-bug3-candidate-amd64 bash
./build_benzhi_docker.sh go-property-rental-bug3-candidate-arm64 linux/arm64
docker run --rm -it --platform linux/arm64 go-property-rental-bug3-candidate-arm64 bash
```

## 题目验证命令
```bash
go test ./scripts/verify -count=1 -run '^TestBug003_BusinessRegression$'
```

## Bug 复现
见 `BUG_REPRO.md`。
