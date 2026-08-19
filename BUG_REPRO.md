# BUG-006 租约版本最新项排序异常

## 复现命令
```bash
go test ./scripts/verify -count=1 -run '^TestBug006_BusinessRegression$'
```

仓储按 `version_no ASC` 返回最旧版本优先；业务服务取出末尾最新版本后又追加回末尾，重排实际为空操作。HTTP 层直接透传该顺序，因此多版本租约的当前版本不会位于结果首位。

本题为 diagnosis；测试模型只调查，不修改生产代码、测试代码或配置。