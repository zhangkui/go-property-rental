退租完成事务功能存在一个需要修复的业务问题：问题表现为完成人和房态恢复破坏事务一致性，正确业务行为是结算完成、审计和房态恢复必须在同一事务一致提交。
请阅读当前分支代码，执行以下命令确认问题：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-019.sh
go build ./...

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过。
