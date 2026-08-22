# go-property-rental 评测说明

## 项目信息

- 项目：`zhangkui/go-property-rental`
- 分支：`test_model_fix11`
- Go 工具链：`golang:1.22`

## 标准命令

```bash
cd /app && GOTOOLCHAIN=local go build ./...
cd /app && GOTOOLCHAIN=local go test ./...
cd /app && GOTOOLCHAIN=local go run ./cmd/api
```

## Docker 构建

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh go-property-rental-bug011-candidate-amd64 linux/amd64
docker run --rm -it --platform linux/amd64 go-property-rental-bug011-candidate-amd64 bash
```

## 题目验证命令

```bash
go test ./scripts/verify -count=1 -run '^TestBug011_BusinessRegression$'
```

本分支保留项目生产代码、运行文件和公开验证测试；轨迹、提示词与内部答案材料不提交到项目仓库。
