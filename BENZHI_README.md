# go-property-rental BUG-018 评测说明

## 项目说明

- 项目：`zhangkui/go-property-rental`
- 分支：`test_model_fix18`
- 题目：BUG-018 租客资料更新字段映射一致性。
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
./build_benzhi_docker.sh go-property-rental-bug018-candidate-amd64 linux/amd64
docker run --rm -it --platform linux/amd64 go-property-rental-bug018-candidate-amd64 bash
```

本次流程未实际运行 Docker、linux/amd64 或 linux/arm64 验证，不声明这些环境通过。

## 题目验证命令

```bash
go test ./scripts/verify -count=1 -run '^TestBug018_BusinessRegression$'
```

## 验证结果

独立 post_pre Session `6d0d0f26-2039-46c3-a955-f19d750ada05` 执行上述命令，退出码为 0：

```text
ok  go-property-rental/scripts/verify  0.394s
```

## Bug 复现

问题现象、触发条件和正确行为见 `BUG_REPRO.md`。
