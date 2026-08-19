# BUG-011 正式交付清单

- [x] `submission-draft.txt`、`verify_result.json`、运行元数据和六项轨迹审计
- [x] `pre_fix.jsonl`、`diagnosis.jsonl` 原始轨迹及上传下载 SHA-256 回验
- [x] Pre 目标测试为 red；正式 Diagnosis 保持工作区零修改
- [x] 诊断命中 service 的统计时点漂移和 repository 的 SLA 阈值错误
- [x] Post `user_query` 不包含验证命令，正式轨迹与文件原文一致
- [x] `verify_cmds` 仅填写对应的 `go test` 命令

结论：BUG-011 的诊断、轨迹审计和交付材料通过检查，可以提交与推送。
