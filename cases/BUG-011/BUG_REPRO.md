# BUG-011 复现说明

- 任务类型：diagnosis
- 功能域：维修 SLA 报表
- 被测分支：`bug11_main` / `test_model_fix11`

## 执行

在项目根目录运行：

```bash
go test ./scripts/verify -count=1 -run '^TestBug011_BusinessRegression$'
```

## 缺陷基线

调用 `ReportService.MaintenanceSLA` 时传入 `2026-08-19 12:00:00 UTC`，数据层实际收到 `2026-08-22 12:00:00 UTC`。公开测试以 `reporting instant drifted` 失败。

## 期望结果

- Service 向 repository 原样传递统计时点。
- 未关闭工单只有创建时间早于统计时点 72 小时时才计入 overdue。
- 正式 diagnosis 只调查原因，不修改生产代码、测试或配置。

## 涉及生产文件

- `internal/service/report.go`
- `internal/repository/mysql/report.go`
