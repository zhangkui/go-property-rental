# 30 题测试协议（私有）

所有题目验证脚本只放在 `cases/BUG-xxx/verification_tests` 和 `cases/BUG-xxx/private_tests`，不复制到项目公开测试目录。

每个题目必须同时满足：

1. 使用独立的 Docker Compose 数据库卷或唯一数据前缀，脚本可重复运行。
2. 先通过 HTTP 登录并取得真实角色令牌，不直接写 MySQL 绕过 API 业务边界。
3. 通过 HTTP 创建完整前置数据，记录每个资源 ID；必要时使用管理员和操作/审批角色分别验证权限。
4. 至少执行一个成功前置动作、一个触发缺陷的动作和一个数据库事实回读动作。
5. 对 HTTP 状态码、稳定错误码/消息、响应关键字段和列表数量做断言；不能只检查命令退出码。
6. 对幂等/并发题重复发送相同请求，检查资源数量、余额、状态历史和审计记录只变化一次。
7. 对事务题在故意失败后回读所有相关实体，确认没有半提交数据；成功路径再确认全部实体一致。
8. 对状态流转题覆盖合法前进、非法回退和重复请求三条路径。
9. 每题的 `verify_cmds.ps1` 必须从仓库根目录执行，输出 `PRE_FIX_RED` 或 `POST_FIX_GREEN` 标记及完整 HTTP 响应摘要。
10. diagnosis 题的验证命令只复现并收集证据，不能修改生产源码；bugfix 题还必须提供 fixed green 命令。

## 固定命令约定

```powershell
docker compose up -d --build
powershell -NoProfile -ExecutionPolicy Bypass -File cases/BUG-xxx/verification_tests/verify_cmds.ps1
```

脚本失败必须以非零退出；成功必须明确输出 `ASSERTIONS_PASSED=<数量>`。每题至少 5 个业务断言，其中至少 2 个跨资源或跨层结果断言。

