# 试标缺陷目录（私有）

本目录仅供校准与验收使用，Gold/Test 模型不得读取。所有题目只允许修改 Go 后端生产代码；前端只能用于人工操作入口和 API 黑盒验证，不设计前端缺陷。

| 编号 | 类型 | 功能域 | 缺陷边界 | 计划涉及生产层 |
|---|---|---|---|---|
| BUG-001 | bugfix | 房源状态 | 房源状态变更服务未拒绝非法回退，handler 仍返回成功 | property service / mysql repository |
| BUG-002 | bugfix | 房源设施 | 替换设施关联时未清理旧关联，详情出现脏设施 | facility service / mysql repository |
| BUG-003 | bugfix | 租客状态 | 停用租客未阻止新租约创建 | tenant service / lease service |
| BUG-004 | diagnosis | 租约重叠 | 时间边界查询漏掉同日交接的重叠租约 | lease service / lease repository |
| BUG-005 | bugfix | 租约续租 | 续租版本金额更新后未同步房态可租日期 | lease service / property repository |
| BUG-006 | diagnosis | 租约版本 | 版本列表排序与当前版本选择不一致 | lease repository / lease service |
| BUG-007 | bugfix | 账单生成 | 幂等命中时重复推进账单计划 | billing service / billing repository |
| BUG-008 | bugfix | 账单滞纳金 | 已结清账单仍可追加滞纳金 | billing service / billing repository |
| BUG-009 | diagnosis | 账单减免 | 减免金额校验使用总额而非未付余额 | billing service / mysql repository |
| BUG-010 | bugfix | 账单部分支付 | 重复收款引用在事务前被接受，产生重复收款 | billing service / payment repository |
| BUG-011 | diagnosis | 收款核销 | 多账单分摊边界金额导致余额出现负数 | billing service / payment repository |
| BUG-012 | bugfix | 押金收取 | 同一业务引用重试时重复增加押金余额 | deposit service / deposit repository |
| BUG-013 | bugfix | 押金扣款 | 扣款未锁定押金余额，两个并发请求均成功 | deposit service / mysql repository |
| BUG-014 | diagnosis | 押金返还 | 已结算租约仍允许独立退款流水 | settlement service / deposit repository |
| BUG-015 | bugfix | 维修工单 | 非派单状态下可直接租客确认 | workorder service / mysql repository |
| BUG-016 | bugfix | 维修材料 | 材料数量更新未同步工单总费用 | workorder service / mysql repository |
| BUG-017 | diagnosis | 维修派单 | 派单给停用用户时错误发生在审计写入之后 | workorder service / auth repository |
| BUG-018 | bugfix | 退租抄表 | 同一租约同一表计重复抄表未拒绝 | settlement service / settlement repository |
| BUG-019 | bugfix | 退租结算 | 押金抵扣与返还金额未在同一锁内计算 | settlement service / deposit repository |
| BUG-020 | diagnosis | 房态恢复 | 结算完成后房态恢复失败但结算已提交 | settlement repository / property repository |
| BUG-021 | bugfix | RBAC | 禁用角色仍可通过旧 token 执行受保护写操作 | auth middleware / security repository |
| BUG-022 | bugfix | RBAC | 用户移除最后一个角色时缓存权限未失效 | rbac service / redis security |
| BUG-023 | diagnosis | 登录限流 | 刷新会话失败未计入登录失败窗口 | auth service / redis security |
| BUG-024 | bugfix | 刷新会话 | 撤销旧刷新令牌后仍可并发刷新出第二会话 | auth service / security repository |
| BUG-025 | bugfix | 审计日志 | 关键状态操作审计失败时业务写入仍提交 | audit service / business service |
| BUG-026 | diagnosis | 通知提醒 | 提醒任务重复执行产生重复通知 | reminder scheduler / notification repository |
| BUG-027 | bugfix | 仪表盘缓存 | Redis 缓存反序列化失败时返回零值而非查询 MySQL | dashboard service / redis dashboard |
| BUG-028 | bugfix | 报表导出 | 白名单排序之外的字段可进入 SQL ORDER BY | report handler / report repository |
| BUG-029 | diagnosis | 审批中心 | 审批步骤跳过当前审核人仍能完成审批 | approval service / approval repository |
| BUG-030 | diagnosis | 账单计划重试 | 失败重试跨租约修改了错误计划的 next_period_start | billing plan service / billing repository |
