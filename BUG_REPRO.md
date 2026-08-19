# BUG-007 账单幂等生成失效

## 复现命令
```bash
go test ./scripts/verify -count=1 -run '^TestBug007_BusinessRegression$'
```

业务层为调用方幂等键追加随机 ID，仓储查询和写入幂等记录又使用不同 scope，导致重复请求无法复用已有账单。公开测试覆盖键原样传递、已有记录复用以及重复请求零新增写入。