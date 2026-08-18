请执行以下命令，验证相关功能的问题是否存在，不能修改任何文件和删除增加任何文件，并报告退出码和关键输出：

docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-020.sh
go build ./...
