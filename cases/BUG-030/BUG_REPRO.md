# BUG-030 复现说明（私有）

- 任务类型：diagnosis
- 功能域：账单计划失败重试
- 被测分支：bug30_main / test_model_fix30
- 前端入口：http://localhost:8080/billing-plans
- 验收账号：admin / Admin123!（仅本地验收管理员；脚本会按场景创建额外账号和数据）

## 前置与重置

1. 检出 bug30_main，确认只存在 scripts/verify/bug-030.sh。
2. 执行 docker compose down -v，删除本题之前的 MySQL/Redis 状态。
3. 执行 docker compose up -d --build，等待 mysql、redis、api 健康。

## 执行

```bash
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-030.sh
go build ./...
```

## 实际结果

- 重试错移失败周期并额外推进计划月份。
- verification test 输出 ASSERTION_FAILED=BUG-030 和 PRE_FIX_RED=BUG-030，退出码为 1。

## 期望结果

- 重试必须针对原失败周期且成功账单不得重复生成或多推进月份。
- 应用本题标准修复后，同一脚本输出 ASSERTIONS_PASSED，退出码为 0。

## 涉及生产文件

- internal/repository/mysql/billing_plan.go
- internal/service/billing_plan.go

## 隔离要求

不得修改或删除测试，不得放宽断言；不得读取 cases、Gold 分支、gold patch、private tests 或其他答案材料。
