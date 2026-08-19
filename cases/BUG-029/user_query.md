审批决定时操作人身份被换掉了。`reviewer-1` 处理 `approval-1`，结果里记录的 reviewer 不是他，测试报 `reviewer identity corrupted`；如果审批单已经指定了处理人，其他人也不应该能代签。请先查清这条审批链路，代码先别改。
