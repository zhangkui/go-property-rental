# BUG-004 租约状态与历史不一致

## 复现命令
```bash
go test ./scripts/verify -count=1 -run '^TestBug004_BusinessRegression$'
```

缺陷版本将合法 `pending` 目标错误改为 `cancelled`。调查还确认仓储更新租约最终状态后，把历史记录的 `to_status` 错写为旧状态，造成最终状态与历史不一致。

本题为 diagnosis；测试模型只调查，不修改生产代码、测试代码或配置。
