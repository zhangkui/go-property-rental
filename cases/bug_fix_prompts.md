# Test 模型 post 提示词汇总（私有）

## BUG-001

房源状态功能存在一个需要修复的跨层业务问题，其中包含两个相互独立且都必须修复的错误行为：一是业务处理层收到非法状态后，会将其转换成合法状态并继续执行更新；二是数据持久化层自身没有执行状态合法性约束，在被直接调用或上游遗漏校验时仍可能将非法状态写入数据库。正确业务行为是业务处理层必须直接拒绝非法状态，不得转换后继续执行；数据持久化层也必须独立拒绝非法状态，不能只依赖上游校验。任何一层收到非法状态都必须返回失败，且数据库原状态保持不变。修复必须完整覆盖这两个错误行为，不能只在单一调用入口增加拦截。

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

同一笔账单收款重试后，对账流水号会变。客户端两次都传 `external-ref`，结果返回的 reference 不一样，后面的分摊记录也有可能找不到本次 payment。这个 reference 是上游用来认同一笔业务的，帮我把重复提交这块修一下，测试先别动。

go test ./scripts/verify -count=1 -run '^TestBug010_BusinessRegression$'

## BUG-011

维修 SLA 报表偶尔会把边界工单算成逾期。调用时传的是 `2026-08-19 12:00:00 UTC`，但测试里记录到数据层收到的统计时间已经偏了，提示 `reporting instant drifted`。想确认一下这个统计时点是在哪儿发生变化的，先不要改代码。

## BUG-012

押金收取重试会被当成一笔新业务。`lease-1` 两次都用 `client-ref` 收 5000，入账后的 reference 却不一样，余额和流水就有重复增加的风险；另外 0 金额不能入账。请处理一下这块重试逻辑，保留现有测试。

## BUG-013

审批撤销有越权情况。`applicant-1` 撤销自己的申请后，审计里的操作人不对；换成另一个用户去撤销，状态和审计也可能照样变化。看下撤销权限和审计记录为什么没对上，只改业务代码就行。

## BUG-014

退租结算时押金会凭空少一部分。押金 5000、抵扣 2000，本来应该退 3000；抵扣为 0 时也应该全额退回，但现在两个场景都对不上。帮我查一下金额从创建结算到计算退款的过程中哪里发生了变化，先不要提交修改。

## BUG-015

维修单还在 `reported` 状态时，租客就能直接点确认关闭，派单和维修过程全被跳过了。测试里还发现 `work-1` 传到后面后被加了额外内容。这个确认流程需要修一下，未完成的工单不能直接关掉，测试文件不要改。

## BUG-016

维修单录材料后金额不对。给 `work-1` 加 3 个 filter，单价 2500，保存出来的数量、单价和 7500 的合计对不上，测试提示 `material values crossed`。请处理材料录入的问题，数量为 0 的请求也不能落库。

## BUG-017

派维修单时负责人会被换掉。`admin-1` 把 `work-1` 指派给 `technician-1`，最后保存的 assignee 却成了管理员，测试报 `assignee and actor crossed`。想查清这两个身份是在哪一步混到一起的，代码先不要改。

## BUG-018

租客资料编辑后字段串了。`tenant-1` 提交的新姓名、手机号、邮箱和证件号都没报错，但回读时电话和邮箱、姓名和证件信息对不上，回归提示 `tenant fields crossed`。把更新资料这块修好，状态和其它字段别受影响。

## BUG-019

完成退租后房源没有恢复成可出租状态，操作记录里的管理员也变成了结算编号。复现是 `admin-1` 完成 `settlement-1`，测试报 `completion actor corrupted`。请处理这条完成流程，租约、房态和操作人要一起保持正确。

## BUG-020

房源列表翻第二页时会漏数据。筛选 `maintenance`，传 `page=2,size=2`，测试期望 offset 是 2，实际拿到的值不对；第一页和空分页参数也有类似偏差。帮忙看看分页为什么越翻越往后跳，修复时保留现有用例。

## BUG-021

用户被停用以后，旧 token 还能继续用。把 `target-1` 设为 disabled 后，原来的 access token 仍能鉴权，refresh token 也能换新会话，而且被清掉的会话有时不是这个用户的。需要把停用后的登录状态处理好，正常 active 用户不要受影响。

## BUG-022

替换用户角色后，旧角色没有完全清掉，反而把操作管理员踢下线了。给 `target-1` 设置 `role-a` 和 `role-b` 后，测试提示撤销会话的对象不是 target。请修一下角色全量替换，空值和重复角色也一起处理掉。

## BUG-023

登录限流会把不同 IP 混在一起。用户名 ` Alice ` 从 `10.0.0.1` 和 `10.0.0.2` 登录时，本来应该各算各的，实际失败次数会互相影响；成功登录后计数也没清干净。想查一下限流键和清理逻辑，先别改代码。

## BUG-024

用户 `user-1` 修改密码后，新密码登录不了，旧密码却还能用，回归报 `stored hash does not accept new password`。输入错误的当前密码时也不能更新任何东西。把改密码流程修好，会话撤销也要针对当前用户。

## BUG-025

审计列表用 `resource=property, actor=admin` 查询时，仓储收到的筛选条件和请求不一样，已有日志就查不到；两个条件都为空时也不能被偷偷补值。请修一下查询筛选，审计详情的正常写入别受影响。

## BUG-026

每日提醒任务在同一天跑两次会生成两次运行记录。8 月 19 日的任务第一次已经成功，第二次不应该再创建新的 run，但回归报 `daily idempotency failed`。先查一下日期和运行记录是怎么对应的，失败重试的情况也别混在一起处理，暂时不要改代码。

## BUG-027

仪表盘缓存偶尔会出现坏 JSON。遇到这种缓存时接口直接返回空汇总，没有去数据库取数据，回归提示 `corrupt cache did not fall back`。缓存读失败后的回源需要修好，正常缓存命中不要改坏。

## BUG-028

租金报表在日期刚好卡边界时会漏记录。查询从 `2026-08-18` 开始时，结束日正好是这一天的租约没有出现在结果里，测试报 `inclusive start boundary shifted`。帮忙查一下 API 到报表查询的日期传递，等于边界的情况要保留。

## BUG-029

审批决定时操作人身份被换掉了。`reviewer-1` 处理 `approval-1`，结果里记录的 reviewer 不是他，测试报 `reviewer identity corrupted`；如果审批单已经指定了处理人，其他人也不应该能代签。请先查清这条审批链路，代码先别改。

## BUG-030

账单计划重试时月份会跑偏。`item-1` 的失败周期是 2026-07-01，重试后应该还是这个周期，但回归发现生成的周期不一致，已有账单的路径还会把计划多推进一个月。成功过的账单项也不该再次生成。先看看重试和生成这两条路径为什么会互相影响，不要修改代码。
