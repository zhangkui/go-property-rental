# 30 题测试协议（私有）

每个 `bugN_main` 分支只提交本题公开验证测试和运行命令，例如 `verification_tests/` 与 `verify_cmds.ps1`；`BUG_REPRO.md`、`user_query.md`、`private_tests/`、`gold_patch.patch`、修复提示词和轨迹材料不得提交到 `bugN_main`、`main` 或 Gold/Test 模型可见分支。30 个缺陷的根因、错误行为和参考修复必须全部位于 Go 后端生产代码；Vue、TypeScript、Nginx、静态资源和前端路由不能作为缺陷文件、根因文件或 Gold 修复文件。

每个题目必须同时满足：

1. 使用独立的 Docker Compose 数据库卷或唯一数据前缀，脚本可重复运行。
2. 先通过 HTTP 登录并取得真实角色令牌，不直接写 MySQL 绕过 API 业务边界。
3. 通过 Go HTTP API 创建完整前置数据，记录每个资源 ID；必要时使用管理员和操作/审批角色分别验证权限。HTTP 仅是黑盒调用入口，不代表把缺陷放到前端。
4. 至少执行一个成功前置动作、一个触发缺陷的动作和一个数据库事实回读动作。
5. 对 HTTP 状态码、稳定错误码/消息、响应关键字段和列表数量做断言；不能只检查命令退出码。
6. 对幂等/并发题重复发送相同请求，检查资源数量、余额、状态历史和审计记录只变化一次。
7. 对事务题在故意失败后回读所有相关实体，确认没有半提交数据；成功路径再确认全部实体一致。
8. 对状态流转题覆盖合法前进、非法回退和重复请求三条路径。
9. 每题的 `verify_cmds.ps1` 必须从仓库根目录执行，输出 `PRE_FIX_RED` 或 `POST_FIX_GREEN` 标记及完整 HTTP 响应摘要。
10. diagnosis 题的验证命令只复现并收集证据，不能修改生产源码；bugfix 题还必须提供 fixed green 命令。每题的根因与正确修复至少涉及两个 Go 生产文件，并跨 repository/service/handler/middleware/platform 中的两个职责层。
11. 每个 `bugN_main` 必须自带且只自带自己的 `verify_cmds.ps1` 和验证测试；执行命令不能引用其他题目目录、主分支测试或未提交的本地文件。复现文档、用户提示词和私有测试在模型运行和分支提交之外单独保存。
12. 只有对应测试在该分支上稳定输出 `PRE_FIX_RED` 且退出码非零，才允许提交 `bugN_main`；红色原因必须来自目标 Go 缺陷，而不是编译失败、Docker 未启动、认证失败或测试数据缺失。

## 固定命令约定

```powershell
docker compose up -d --build
powershell -NoProfile -ExecutionPolicy Bypass -File cases/BUG-xxx/verification_tests/verify_cmds.ps1
```

脚本失败必须以非零退出；成功必须明确输出 `ASSERTIONS_PASSED=<数量>`。每题至少 5 个业务断言，其中至少 2 个跨资源或跨层结果断言。
