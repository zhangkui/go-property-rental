# go-property-rental BUG-027 评测说明

## 项目说明

- 项目：`zhangkui/go-property-rental`
- 分支：`test_model_fix27`
- 题目：BUG-027 仪表盘损坏缓存回退数据库。
- Go 工具链：`golang:1.22`

## 标准构建、运行和测试命令

进入评测容器后执行：

```bash
cd /app && GOTOOLCHAIN=local go build ./...
cd /app && GOTOOLCHAIN=local go test ./...
cd /app && GOTOOLCHAIN=local go run ./cmd/api
```

## Docker 构建和进入容器

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh go-property-rental-bug027-candidate-amd64 linux/amd64
docker run --rm -it --platform linux/amd64 go-property-rental-bug027-candidate-amd64 bash
```

本次流程未实际运行 Docker、linux/amd64 或 linux/arm64 验证，不声明这些环境通过。

## 题目验证命令

在项目根目录执行：

```bash
go test ./scripts/verify -count=1 -run '^TestBug027_BusinessRegression$'
```

## 验证结果

独立 post_pre Session `fe620f9a-7cdb-4c21-80f9-59111d2ac2ec` 执行上述命令，退出码为 0：

```text
ok  go-property-rental/scripts/verify  0.340s
```

## Bug 复现

问题现象、触发条件和正确行为见 `BUG_REPRO.md`。