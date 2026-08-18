#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-011;source scripts/verify/testlib.sh;login_admin
p=$(new_id);recent=$(new_id);overdue=$(new_id);closed=$(new_id);room=${p:0:8}
sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG011','$room','occupied','2026-01-01');INSERT INTO work_orders(id,property_id,description,status,created_at,updated_at) VALUES('$recent','$p','48 hour open','reported',UTC_TIMESTAMP()-INTERVAL 48 HOUR,UTC_TIMESTAMP()-INTERVAL 48 HOUR),('$overdue','$p','96 hour open','assigned',UTC_TIMESTAMP()-INTERVAL 96 HOUR,UTC_TIMESTAMP()-INTERVAL 96 HOUR),('$closed','$p','resolved order','closed',UTC_TIMESTAMP()-INTERVAL 100 HOUR,UTC_TIMESTAMP()-INTERVAL 10 HOUR);"
request GET "/api/reports/maintenance-sla?property_id=$p";body=$HTTP_BODY
total=$(json_number Total "$body");open=$(json_number Open "$body");completed=$(json_number Completed "$body");overdue_count=$(json_number Overdue "$body");average=$(json_number AverageResolutionHours "$body")
[[ $HTTP_STATUS == 200 ]]||fail "maintenance SLA expected HTTP 200 got $HTTP_STATUS body=$body"
[[ $total == 3 ]]||fail "maintenance SLA total expected 3 got $total"
[[ $open == 2 ]]||fail "maintenance SLA open expected 2 got $open"
[[ $completed == 1 ]]||fail "maintenance SLA completed expected 1 got $completed"
[[ $overdue_count == 1 ]]||fail "only 96-hour open order should be overdue got $overdue_count"
[[ $average == 90 ]]||fail "closed order average resolution expected 90 hours got $average"
finish 6
