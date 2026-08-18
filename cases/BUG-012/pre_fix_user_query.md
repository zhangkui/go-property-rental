你正在执行 BUG-012 的修复前验证。不得修改任何文件。请依次、逐条实际执行以下命令并报告每条命令的真实退出码和关键输出：

docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-012.sh
go build ./...
