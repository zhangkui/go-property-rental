# BUG-003 复现说明（私有）

- 任务类型：bugfix
- 功能域：租客状态与租约准入
- 被测分支：bug3_main / test_model_fix3
- 前端入口：http://localhost:8080/tenants、/leases
- 验收账号：admin / Admin123!（仅本地验收管理员；脚本会按场景创建额外账号和数据）

## 前置与重置

1. 检出 bug3_main，确认只存在 scripts/verify/bug-003.sh。
2. 执行 docker compose down -v，删除本题之前的 MySQL/Redis 状态。
3. 执行 docker compose up -d --build，等待 mysql、redis、api 健康。

## 执行

```bash
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-003.sh
go build ./...
```

## 实际结果

- disabled 租客被错误映射并可创建租约。
- verification test 输出 ASSERTION_FAILED=BUG-003 和 PRE_FIX_RED=BUG-003，退出码为 1。

## 期望结果

- 停用租客必须保持 disabled 且不能进入新租约。
- 应用本题标准修复后，同一脚本输出 ASSERTIONS_PASSED，退出码为 0。

## 涉及生产文件

- internal/repository/mysql/tenant.go
- internal/service/lease.go

## 隔离要求

不得修改或删除测试，不得放宽断言；不得读取 cases、Gold 分支、gold patch、private tests 或其他答案材料。
