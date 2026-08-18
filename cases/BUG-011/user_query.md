维修 SLA 报表功能存在一个需要调查的业务异常：问题表现为72 小时边界被扭曲为错误统计时点和 24 小时阈值，正确业务行为是统计时点不得漂移且仅超过 72 小时的未结工单算逾期。
请阅读当前分支代码，并执行以下命令收集问题证据：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-011.sh
go build ./...

请先分析相关 Go 后端生产代码、数据访问和 HTTP 调用链，定位具体文件、函数或方法、触发路径以及跨层失效机制。先不要改目标仓库代码，全程不得修改目标仓库中的生产代码、测试代码或配置。
请调查并解释该异常的根本原因，给出可核查的调查证据。
