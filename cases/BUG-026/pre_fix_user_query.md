请执行以下命令，验证相关功能的问题是否存在，不能修改任何文件和删除增加任何文件，并报告退出码和关键输出：

go test ./scripts/verify -count=1 -run '^TestBug026_BusinessRegression$'
