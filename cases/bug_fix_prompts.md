# Test 模型 post 提示词汇总（私有）

## BUG-001

房源状态功能存在一个需要修复的业务问题：问题表现为非法状态被当作成功写入，正确业务行为是非法状态必须返回失败且数据库状态保持不变。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-001.sh
go build ./...

## BUG-002

房源设施功能存在一个需要修复的业务问题：问题表现为替换设施后旧关联仍残留，正确业务行为是替换后关联集合必须与请求完全一致且无脏数据。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-002.sh
go build ./...

## BUG-003

租客状态与租约准入功能存在一个需要修复的业务问题：问题表现为disabled 租客被错误映射并可创建租约，正确业务行为是停用租客必须保持 disabled 且不能进入新租约。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-003.sh
go build ./...

## BUG-004

租约状态流转功能存在一个需要调查的业务异常：问题表现为租约当前状态与状态历史记录不一致，正确业务行为是状态必须按合法顺序流转且历史记录与最终状态一致。
请阅读当前分支代码，并执行以下命令收集问题证据：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-004.sh
go build ./...

请先分析相关 Go 后端生产代码、数据访问和 HTTP 调用链，定位具体文件、函数或方法、触发路径以及跨层失效机制。先不要改目标仓库代码，全程不得修改目标仓库中的生产代码、测试代码或配置。
请调查并解释该异常的根本原因，给出可核查的调查证据。

## BUG-005

租约续租功能存在一个需要修复的业务问题：问题表现为续租租金和押金跨层错位，正确业务行为是续租版本必须分别保存正确租金与押金。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-005.sh
go build ./...

## BUG-006

租约版本功能存在一个需要调查的业务异常：问题表现为接口返回最旧合同版本作为当前版本，正确业务行为是当前版本必须按最新版本号和生效顺序返回。
请阅读当前分支代码，并执行以下命令收集问题证据：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-006.sh
go build ./...

请先分析相关 Go 后端生产代码、数据访问和 HTTP 调用链，定位具体文件、函数或方法、触发路径以及跨层失效机制。先不要改目标仓库代码，全程不得修改目标仓库中的生产代码、测试代码或配置。
请调查并解释该异常的根本原因，给出可核查的调查证据。

## BUG-007

账单幂等生成功能存在一个需要修复的业务问题：问题表现为相同幂等键重复生成账单，正确业务行为是重复请求必须返回同一结果且不得新增第二张账单。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-007.sh
go build ./...

## BUG-008

账单滞纳金功能存在一个需要修复的业务问题：问题表现为已结清账单仍增加滞纳金和余额，正确业务行为是已结清账单不得再次产生滞纳金或余额变化。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-008.sh
go build ./...

## BUG-009

通知已读权限功能存在一个需要调查的业务异常：问题表现为用户可以把其他用户的通知标记为已读，正确业务行为是只能由通知所属用户或授权管理员执行已读操作。
请阅读当前分支代码，并执行以下命令收集问题证据：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-009.sh
go build ./...

请先分析相关 Go 后端生产代码、数据访问和 HTTP 调用链，定位具体文件、函数或方法、触发路径以及跨层失效机制。先不要改目标仓库代码，全程不得修改目标仓库中的生产代码、测试代码或配置。
请调查并解释该异常的根本原因，给出可核查的调查证据。

## BUG-010

收款幂等与核销功能存在一个需要修复的业务问题：问题表现为重复收款引用产生重复收款和核销，正确业务行为是相同业务引用只能成功一次且账单余额守恒。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-010.sh
go build ./...

## BUG-011

维修 SLA 报表功能存在一个需要调查的业务异常：问题表现为72 小时边界被扭曲为错误统计时点和 24 小时阈值，正确业务行为是统计时点不得漂移且仅超过 72 小时的未结工单算逾期。
请阅读当前分支代码，并执行以下命令收集问题证据：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-011.sh
go build ./...

请先分析相关 Go 后端生产代码、数据访问和 HTTP 调用链，定位具体文件、函数或方法、触发路径以及跨层失效机制。先不要改目标仓库代码，全程不得修改目标仓库中的生产代码、测试代码或配置。
请调查并解释该异常的根本原因，给出可核查的调查证据。

## BUG-012

押金收取幂等功能存在一个需要修复的业务问题：问题表现为重复押金引用重复增加余额和流水，正确业务行为是相同业务引用不得重复增加押金余额或流水。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-012.sh
go build ./...

## BUG-013

审批撤销权限与审计功能存在一个需要修复的业务问题：问题表现为非申请人可撤销且审计 actor 写错，正确业务行为是仅申请人可撤销，审批状态与审计 actor 必须一致。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-013.sh
go build ./...

## BUG-014

退租押金结算功能存在一个需要调查的业务异常：问题表现为押金抵扣和返还金额跨层漂移，正确业务行为是费用、抵扣、返还与押金流水必须守恒。
请阅读当前分支代码，并执行以下命令收集问题证据：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-014.sh
go build ./...

请先分析相关 Go 后端生产代码、数据访问和 HTTP 调用链，定位具体文件、函数或方法、触发路径以及跨层失效机制。先不要改目标仓库代码，全程不得修改目标仓库中的生产代码、测试代码或配置。
请调查并解释该异常的根本原因，给出可核查的调查证据。

## BUG-015

维修租客确认功能存在一个需要修复的业务问题：问题表现为reported 工单可跳过派单和维修直接确认，正确业务行为是必须完成合法状态链后租客才能确认。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-015.sh
go build ./...

## BUG-016

维修材料费用功能存在一个需要修复的业务问题：问题表现为材料数量、单价和总价被交叉写错，正确业务行为是材料明细与总费用必须按数量乘单价准确汇总。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-016.sh
go build ./...

## BUG-017

维修派单功能存在一个需要调查的业务异常：问题表现为指定维修人员被当前操作人覆盖，正确业务行为是派单对象必须保持请求中的维修人员并记录操作人。
请阅读当前分支代码，并执行以下命令收集问题证据：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-017.sh
go build ./...

请先分析相关 Go 后端生产代码、数据访问和 HTTP 调用链，定位具体文件、函数或方法、触发路径以及跨层失效机制。先不要改目标仓库代码，全程不得修改目标仓库中的生产代码、测试代码或配置。
请调查并解释该异常的根本原因，给出可核查的调查证据。

## BUG-018

租客资料功能存在一个需要修复的业务问题：问题表现为姓名、电话和证件字段错位保存，正确业务行为是所有租客字段必须按 API 契约准确持久化。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-018.sh
go build ./...

## BUG-019

退租完成事务功能存在一个需要修复的业务问题：问题表现为完成人和房态恢复破坏事务一致性，正确业务行为是结算完成、审计和房态恢复必须在同一事务一致提交。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-019.sh
go build ./...

## BUG-020

房源分页功能存在一个需要修复的业务问题：问题表现为offset 在服务层和仓储层重复应用导致跳页，正确业务行为是分页偏移只能应用一次且相邻页连续无遗漏。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-020.sh
go build ./...

## BUG-021

用户停用与会话功能存在一个需要修复的业务问题：问题表现为停用用户的访问令牌和刷新令牌仍有效，正确业务行为是停用必须撤销全部会话并阻止继续访问和刷新。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-021.sh
go build ./...

## BUG-022

用户角色授权功能存在一个需要修复的业务问题：问题表现为替换角色后旧角色残留且旧令牌权限未失效，正确业务行为是角色集合必须完整替换并使旧权限会话失效。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-022.sh
go build ./...

## BUG-023

登录失败限流功能存在一个需要调查的业务异常：问题表现为限流丢失 IP 维度且成功登录清理错误 Redis 键，正确业务行为是同用户不同 IP 独立计数，成功登录只重置对应键，第六次失败限流。
请阅读当前分支代码，并执行以下命令收集问题证据：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-023.sh
go build ./...

请先分析相关 Go 后端生产代码、数据访问和 HTTP 调用链，定位具体文件、函数或方法、触发路径以及跨层失效机制。先不要改目标仓库代码，全程不得修改目标仓库中的生产代码、测试代码或配置。
请调查并解释该异常的根本原因，给出可核查的调查证据。

## BUG-024

修改密码功能存在一个需要修复的业务问题：问题表现为明文密码写入错误用户且会话撤销目标错误，正确业务行为是必须 bcrypt 保存到当前用户并撤销该用户既有会话。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-024.sh
go build ./...

## BUG-025

审计日志查询功能存在一个需要修复的业务问题：问题表现为resource 前缀在服务层和仓储层重复添加，正确业务行为是resource 过滤值必须原样传递并返回匹配日志。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-025.sh
go build ./...

## BUG-026

通知任务幂等功能存在一个需要调查的业务异常：问题表现为每日任务键随机且仓储幂等检查失效，正确业务行为是同一任务周期只能生成一次通知，失败可重试但成功不重复。
请阅读当前分支代码，并执行以下命令收集问题证据：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-026.sh
go build ./...

请先分析相关 Go 后端生产代码、数据访问和 HTTP 调用链，定位具体文件、函数或方法、触发路径以及跨层失效机制。先不要改目标仓库代码，全程不得修改目标仓库中的生产代码、测试代码或配置。
请调查并解释该异常的根本原因，给出可核查的调查证据。

## BUG-027

仪表盘缓存降级功能存在一个需要修复的业务问题：问题表现为Redis 缓存损坏时返回零值，正确业务行为是缓存损坏或不可解析时必须回退 MySQL 事实数据。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-027.sh
go build ./...

## BUG-028

租金报表日期边界功能存在一个需要修复的业务问题：问题表现为租期结束日等于查询开始日时被排除，正确业务行为是日期范围必须包含相等边界且 API/CSV 结果一致。

请只修改 Go 后端生产代码，定位并修复这个跨层业务问题。不得新增、删除或修改任何测试文件，不得跳过测试或放宽测试断言，也不得修改 Docker 验证脚本。只修复提到的业务问题，不扩展修复其他无关问题。
修复完成后，需要执行以下命令验证，保证命令获取结果全绿，并确认相关功能测试、go build ./... 和合法业务场景全部通过：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-028.sh
go build ./...

## BUG-029

审批人权限功能存在一个需要调查的业务异常：问题表现为非当前审批人可以完成审批，正确业务行为是只有当前审批人可审批并留下正确状态和审计记录。
请阅读当前分支代码，并执行以下命令收集问题证据：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-029.sh
go build ./...

请先分析相关 Go 后端生产代码、数据访问和 HTTP 调用链，定位具体文件、函数或方法、触发路径以及跨层失效机制。先不要改目标仓库代码，全程不得修改目标仓库中的生产代码、测试代码或配置。
请调查并解释该异常的根本原因，给出可核查的调查证据。

## BUG-030

账单计划失败重试功能存在一个需要调查的业务异常：问题表现为重试错移失败周期并额外推进计划月份，正确业务行为是重试必须针对原失败周期且成功账单不得重复生成或多推进月份。
请阅读当前分支代码，并执行以下命令收集问题证据：
docker compose down -v
docker compose up -d --build
docker compose -f docker-compose.yml -f docker-compose.verify.yml run --rm verifier scripts/verify/bug-030.sh
go build ./...

请先分析相关 Go 后端生产代码、数据访问和 HTTP 调用链，定位具体文件、函数或方法、触发路径以及跨层失效机制。先不要改目标仓库代码，全程不得修改目标仓库中的生产代码、测试代码或配置。
请调查并解释该异常的根本原因，给出可核查的调查证据。

