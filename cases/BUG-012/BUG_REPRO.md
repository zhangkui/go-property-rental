# BUG-012 复现说明

- 任务类型：bugfix
- 功能域：押金收取幂等
- 被测分支：`bug12_main` / `test_model_fix12`

## 执行

在项目根目录运行：

```bash
go test ./scripts/verify -count=1 -run '^TestBug012_BusinessRegression$'
```

## 缺陷基线

`DepositService.Transaction` 收到 `client-ref` 后返回带随机 UUID 后缀的 reference。重复请求无法保持同一业务引用，公开测试以 `reference rewritten` 失败。

## 期望结果

- `DepositLedger.Reference` 原样保存客户端业务引用。
- 相同租约的重复调用保持相同 reference。
- 金额为 0 的押金操作返回错误且不进入仓储。
- 不修改、删除或放宽公开测试断言。

## 涉及生产文件

- `internal/service/deposit.go`
- `internal/repository/mysql/deposit.go`
