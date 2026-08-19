完成退租后房源没有恢复成可出租状态，操作记录里的管理员也变成了结算编号。复现是 `admin-1` 完成 `settlement-1`，测试报 `completion actor corrupted`。请处理这条完成流程，租约、房态和操作人要一起保持正确。
