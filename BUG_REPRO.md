# BUG-005 租约续租金额跨层错位

## 复现命令
```bash
go test ./scripts/verify -count=1 -run '^TestBug005_BusinessRegression$'
```

缺陷版本在业务层构造续租版本时将 `Deposit` 错取为租金，并在 MySQL 仓储更新主租约时将押金同时绑定到 `monthly_rent` 与 `deposit`。测试同时验证续租版本参数、主租约更新参数以及结束日期不增长时不得进入仓储。

正确修复分别保留调用方传入的租金和押金，并确保续租版本与主租约记录一致。