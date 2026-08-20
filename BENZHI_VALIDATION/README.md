# BENZHI_VALIDATION

本目录保存 BUG-027 的真实验证结果和环境说明。

- 验证环境：Go 1.26.1，Windows/amd64。
- 未运行 Docker、linux/amd64 或 linux/arm64 验证。
- post_pre 原始 Session：`fe620f9a-7cdb-4c21-80f9-59111d2ac2ec`。
- 固定验证命令：`go test ./scripts/verify -count=1 -run '^TestBug027_BusinessRegression$'`。
- 结果：退出码 0，green。