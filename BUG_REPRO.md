# BUG-001 房源非法状态写入

## 问题现象

提交非法房源状态时，业务层会将非法值静默转换为 `available` 并继续更新；持久化层也没有独立验证状态集合。请求因此可能返回成功、覆盖数据库原状态，并产生成功审计记录。

## 触发与验证

在项目根目录执行：

```bash
go test ./scripts/verify -count=1 -run '^TestBug001_BusinessRegression$'
```

缺陷版本会在非法状态相关子测试中失败。

## 正确行为

- 非法状态必须返回错误。
- 数据库原状态必须保持不变。
- 被拒绝的状态更新不能产生成功审计记录。
- 合法的 `maintenance` 状态仍能成功持久化并产生正常审计记录。

## 修复后结果

同一验证命令应当通过：

```text
ok  go-property-rental/scripts/verify
```
