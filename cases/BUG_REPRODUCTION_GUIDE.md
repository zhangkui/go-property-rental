# BUG 复现总指南（私有）

本文件以及各题 private patch、user query 仅用于校准和验收，不得复制到模型可见分支。

| 编号 | 类型 | 功能域 | 前端路径 | 验证脚本 |
|---|---|---|---|---|
| BUG-001 | bugfix | 房源状态 | /properties | scripts/verify/bug-001.sh |
| BUG-002 | bugfix | 房源设施 | /facilities | scripts/verify/bug-002.sh |
| BUG-003 | bugfix | 租客状态与租约准入 | /tenants、/leases | scripts/verify/bug-003.sh |
| BUG-004 | diagnosis | 租约状态流转 | /leases | scripts/verify/bug-004.sh |
| BUG-005 | bugfix | 租约续租 | /leases | scripts/verify/bug-005.sh |
| BUG-006 | diagnosis | 租约版本 | /leases | scripts/verify/bug-006.sh |
| BUG-007 | bugfix | 账单幂等生成 | /bills | scripts/verify/bug-007.sh |
| BUG-008 | bugfix | 账单滞纳金 | /bills | scripts/verify/bug-008.sh |
| BUG-009 | diagnosis | 通知已读权限 | /notifications | scripts/verify/bug-009.sh |
| BUG-010 | bugfix | 收款幂等与核销 | /bills | scripts/verify/bug-010.sh |
| BUG-011 | diagnosis | 维修 SLA 报表 | /reports | scripts/verify/bug-011.sh |
| BUG-012 | bugfix | 押金收取幂等 | /deposits | scripts/verify/bug-012.sh |
| BUG-013 | bugfix | 审批撤销权限与审计 | /approvals | scripts/verify/bug-013.sh |
| BUG-014 | diagnosis | 退租押金结算 | /settlements | scripts/verify/bug-014.sh |
| BUG-015 | bugfix | 维修租客确认 | /work-orders | scripts/verify/bug-015.sh |
| BUG-016 | bugfix | 维修材料费用 | /work-orders | scripts/verify/bug-016.sh |
| BUG-017 | diagnosis | 维修派单 | /work-orders | scripts/verify/bug-017.sh |
| BUG-018 | bugfix | 租客资料 | /tenants | scripts/verify/bug-018.sh |
| BUG-019 | bugfix | 退租完成事务 | /settlements | scripts/verify/bug-019.sh |
| BUG-020 | bugfix | 房源分页 | /properties | scripts/verify/bug-020.sh |
| BUG-021 | bugfix | 用户停用与会话 | /users | scripts/verify/bug-021.sh |
| BUG-022 | bugfix | 用户角色授权 | /users、/roles | scripts/verify/bug-022.sh |
| BUG-023 | diagnosis | 登录失败限流 | /login | scripts/verify/bug-023.sh |
| BUG-024 | bugfix | 修改密码 | /profile | scripts/verify/bug-024.sh |
| BUG-025 | bugfix | 审计日志查询 | /audits | scripts/verify/bug-025.sh |
| BUG-026 | diagnosis | 通知任务幂等 | /notifications | scripts/verify/bug-026.sh |
| BUG-027 | bugfix | 仪表盘缓存降级 | / | scripts/verify/bug-027.sh |
| BUG-028 | bugfix | 租金报表日期边界 | /reports | scripts/verify/bug-028.sh |
| BUG-029 | diagnosis | 审批人权限 | /approvals | scripts/verify/bug-029.sh |
| BUG-030 | diagnosis | 账单计划失败重试 | /billing-plans | scripts/verify/bug-030.sh |
