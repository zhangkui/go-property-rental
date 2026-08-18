# BUG-028 复现说明（私有）

- 任务类型：bugfix
- 功能域：租金报表日期边界
- 被测分支：bug28_main / test_model_fix28
- 前端入口：http://localhost:8080/reports
- 验收账号：admin / Admin123!（仅本地验收管理员；脚本会按场景创建额外账号和数据）

## 前置与重置

1. 检出 bug28_main，确认只存在 scripts/verify/bug-028.sh。
2. 执行 docker compose down -v，删除本题之前的 MySQL/Redis 状态。
3. 执行 docker compose up -d --build，等待 mysql、redis、api 健康。

## 执行

```bash
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-028.sh
go build ./...
```

## 实际结果

- 租期结束日等于查询开始日时被排除。
- verification test 输出 ASSERTION_FAILED=BUG-028 和 PRE_FIX_RED=BUG-028，退出码为 1。

## 期望结果

- 日期范围必须包含相等边界且 API/CSV 结果一致。
- 应用本题标准修复后，同一脚本输出 ASSERTIONS_PASSED，退出码为 0。

## 涉及生产文件

- internal/repository/mysql/report.go
- internal/service/report.go

## 隔离要求

不得修改或删除测试，不得放宽断言；不得读取 cases、Gold 分支、gold patch、private tests 或其他答案材料。
