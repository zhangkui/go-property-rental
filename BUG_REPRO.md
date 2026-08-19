# BUG-003 disabled 租客错误进入新租约

## 复现命令
```bash
go test ./scripts/verify -count=1 -run '^TestBug003_BusinessRegression$'
```

缺陷版本返回 `disabled tenant was allowed to create a lease`。正确行为是租客查询保持数据库真实 `disabled` 状态，disabled 及其他不可租约状态不得创建新租约且不得调用写入；active 租客仍可正常创建。

涉及文件：
- `internal/repository/mysql/tenant.go`
- `internal/service/lease.go`
