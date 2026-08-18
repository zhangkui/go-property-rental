审批撤销权限与审计功能存在一个需要修复的业务问题。
问题表现：非申请人可撤销且审计 actor 写错。
正确业务行为：仅申请人可撤销，审批状态与审计 actor 必须一致。
请阅读当前分支代码，先实际执行以下命令确认问题：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-013.sh
go build ./...

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。
修复完成后，必须再次逐条执行下面完全相同的命令，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过。
