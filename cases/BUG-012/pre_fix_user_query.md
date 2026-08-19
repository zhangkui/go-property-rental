请只执行以下命令，验证相关功能的问题是否存在。不要修改、创建或删除任何文件，最后报告关键输出：

go test ./scripts/verify -count=1 -run '^TestBug012_BusinessRegression$'
