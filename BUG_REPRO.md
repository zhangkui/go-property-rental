# BUG-008 已结清账单滞纳金异常

## 复现命令
```bash
go test ./scripts/verify -count=1 -run '^TestBug008_BusinessRegression$'
```

业务层原先以账单原始金额计算滞纳金，已全额支付账单仍会产生 penalty；Repository 原先允许 paid 账单的 penalty 调整，导致绕过业务层直接写入余额。

正确行为是已结清账单不产生滞纳金、不新增调整记录、不改变余额；逾期且有未结余额的账单按未结余额计算费用。