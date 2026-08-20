# BUG-026

复现与验收只执行仓库内 Go 测试：

```text
go test ./scripts/verify -count=1 -run '^TestBug026_BusinessRegression$'
```

不要使用 Docker、shell 脚本或修改测试文件。
