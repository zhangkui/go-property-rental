# 试标缺陷目录（私有）

本目录仅供校准与验收使用，Gold/Test 模型不得读取。所有题目只允许修改 Go 后端生产代码；前端只能用于人工操作入口和 API 黑盒验证，不设计前端缺陷。

| 编号 | 类型 | 功能域 | 缺陷边界 | 计划涉及生产层 |
|---|---|---|---|---|
| BUG-001 | bugfix | 房源状态 | 非法房源状态被规范化后持久化，接口错误返回成功 | property service / mysql repository |
| BUG-002 | bugfix | 房源设施 | 替换设施关联时未清理旧关联，详情出现脏设施 | facility service / mysql repository |
| BUG-003 | bugfix | 租客状态 | 数据层把 disabled 映射成 suspended，租约服务绕过停用校验 | tenant repository / lease service |
| BUG-004 | diagnosis | 租约状态流转 | 服务层状态流转与仓储状态历史写入不一致，当前状态和历史状态分叉 | lease service / lease repository |
| BUG-005 | bugfix | 租约续租 | 续租租金与押金在服务层和持久化层发生交叉错配 | lease service / lease repository |
| BUG-006 | diagnosis | 租约版本 | 租约版本按最旧优先返回，当前版本选择错误 | lease repository / lease service |
| BUG-007 | bugfix | 账单生成 | 手工生成账单的幂等键失效，重复请求生成重复账单 | billing service / billing repository |
| BUG-008 | bugfix | 账单滞纳金 | 已结清账单仍被追加滞纳金并改变余额 | billing service / billing repository |
| BUG-009 | diagnosis | 通知已读 | 服务层忽略当前用户身份且仓储层不校验通知归属，用户可标记他人通知已读 | notification service / notification repository |
| BUG-010 | bugfix | 账单收款 | 重复收款引用被接受，产生重复收款与重复核销 | billing service / billing repository |
| BUG-011 | diagnosis | 维修 SLA 报表 | 服务层推进统计时点且仓储层把 72 小时阈值缩短为 24 小时，边界工单被误判逾期 | report service / report repository |
| BUG-012 | bugfix | 押金收取 | 重复押金业务引用被接受，余额和流水重复增加 | deposit service / deposit repository |
| BUG-013 | bugfix | 审批撤销 | 服务层未校验申请人且仓储层写错撤销人，越权撤销与审计 actor 同时失真 | approval service / approval repository / audit repository |
| BUG-014 | diagnosis | 退租结算 | 结算中的押金抵扣与返还金额跨层漂移 | settlement service / settlement repository |
| BUG-015 | bugfix | 维修确认 | reported 工单可跳过派单和处理直接由租客确认 | workorder service / workorder repository |
| BUG-016 | bugfix | 维修材料 | 材料数量、单价和总费用计算被交叉写错 | workorder service / workorder repository |
| BUG-017 | diagnosis | 维修派单 | 指定的维修人员被当前操作人 ID 覆盖 | workorder service / workorder repository |
| BUG-018 | bugfix | 租客资料 | 租客姓名、电话、证件等资料字段跨层错位映射 | tenant service / tenant repository |
| BUG-019 | bugfix | 退租完成 | 完成人与房态恢复写入破坏同一事务的一致性 | settlement service / settlement repository |
| BUG-020 | bugfix | 房源分页 | 服务层和数据层重复应用 offset，分页结果跳页 | property service / mysql repository |
| BUG-021 | bugfix | 用户停用 | 用户停用后旧访问令牌和刷新令牌仍然有效 | auth/rbac service / security repository |
| BUG-022 | bugfix | 用户授权 | 替换用户角色时旧角色残留且旧令牌权限未失效 | rbac service / security repository |
| BUG-023 | diagnosis | 登录限流 | 限流键丢失 IP 维度且成功登录清理了错误键 | auth service / redis security repository |
| BUG-024 | bugfix | 修改密码 | 服务层保存明文新密码且仓储层更新错误用户，密码变更和会话撤销目标不一致 | auth service / security repository |
| BUG-025 | bugfix | 审计查询 | resource 过滤条件被重复添加命名空间前缀 | audit service / audit repository |
| BUG-026 | diagnosis | 通知提醒 | 每日任务键随机化且仓储幂等检查被移除，重复生成提醒 | notification service / notification repository |
| BUG-027 | bugfix | 仪表盘缓存 | Redis 缓存损坏时返回零值，不回退 MySQL 事实数据 | dashboard service / redis dashboard repository |
| BUG-028 | bugfix | 租金报表 | 租期结束日边界在 API 与 CSV 查询中被排除 | report service / report repository |
| BUG-029 | diagnosis | 审批中心 | 当前审批人校验被绕过，非当前审核人可完成审批 | approval service / approval repository |
| BUG-030 | diagnosis | 账单计划重试 | 重试把失败周期错移一月，计划又额外推进一月 | billing plan service / billing plan repository |
