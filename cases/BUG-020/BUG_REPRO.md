# BUG-020 复现说明（私有）

- 任务类型：bugfix
- 功能域：房源分页
- 被测分支：bug20_main / test_model_fix20
- 前端入口：http://localhost:8080/properties
- 验收账号：admin / Admin123!（仅本地验收管理员；脚本会按场景创建额外账号和数据）

## 前置与重置

1. 检出 bug20_main，确认只存在 scripts/verify/bug-020.sh。
2. 执行 docker compose down -v，删除本题之前的 MySQL/Redis 状态。
3. 执行 docker compose up -d --build，等待 mysql、redis、api 健康。

## 执行

```bash
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-020.sh
go build ./...
```

## 实际结果

- offset 在服务层和仓储层重复应用导致跳页。
- verification test 输出 ASSERTION_FAILED=BUG-020 和 PRE_FIX_RED=BUG-020，退出码为 1。

## 期望结果

- 分页偏移只能应用一次且相邻页连续无遗漏。
- 应用本题标准修复后，同一脚本输出 ASSERTIONS_PASSED，退出码为 0。

## 涉及生产文件

- internal/repository/mysql/repositories.go
- internal/service/property.go

## 隔离要求

不得修改或删除测试，不得放宽断言；不得读取 cases、Gold 分支、gold patch、private tests 或其他答案材料。
