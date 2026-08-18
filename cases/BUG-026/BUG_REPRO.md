# BUG-026 复现说明（私有）

- 任务类型：diagnosis
- 功能域：通知任务幂等
- 被测分支：bug26_main / test_model_fix26
- 前端入口：http://localhost:8080/notifications
- 验收账号：admin / Admin123!（仅本地验收管理员；脚本会按场景创建额外账号和数据）

## 前置与重置

1. 检出 bug26_main，确认只存在 scripts/verify/bug-026.sh。
2. 执行 docker compose down -v，删除本题之前的 MySQL/Redis 状态。
3. 执行 docker compose up -d --build，等待 mysql、redis、api 健康。

## 执行

```bash
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-026.sh
go build ./...
```

## 实际结果

- 每日任务键随机且仓储幂等检查失效。
- verification test 输出 ASSERTION_FAILED=BUG-026 和 PRE_FIX_RED=BUG-026，退出码为 1。

## 期望结果

- 同一任务周期只能生成一次通知，失败可重试但成功不重复。
- 应用本题标准修复后，同一脚本输出 ASSERTIONS_PASSED，退出码为 0。

## 涉及生产文件

- internal/repository/mysql/notification.go
- internal/service/notification.go

## 隔离要求

不得修改或删除测试，不得放宽断言；不得读取 cases、Gold 分支、gold patch、private tests 或其他答案材料。
