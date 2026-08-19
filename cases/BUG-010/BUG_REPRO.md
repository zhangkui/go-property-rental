# BUG-010 复现说明

- 任务类型：bugfix
- 功能域：收款幂等与核销
- 被测分支：`bug10_main` / `test_model_fix10`

## 执行

在项目根目录运行：

```bash
go test ./scripts/verify -count=1 -run '^TestBug010_BusinessRegression$'
```

## 缺陷基线

相同账单收款请求重复使用 `external-ref` 时，`BillingService.Pay` 为每次调用生成不同的 `Payment.Reference`。测试在 `bug-010_test.go:44` 处失败，报告客户端 reference 被改写。

## 期望结果

- 返回的 `Payment.Reference` 保持客户端传入的业务引用。
- 重试请求继续使用同一业务引用，不能被内部 ID 改写成不同流水号。
- 每条分摊记录的 `PaymentID` 指向本次返回的 payment ID。
- 不修改、删除或放宽公开测试断言。

## 涉及生产文件

- `internal/service/billing.go`
- `internal/repository/mysql/billing.go`
