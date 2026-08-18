#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-015;source scripts/verify/testlib.sh;login_admin
p=$(new_id);room=${p:0:8};sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG015','$room','available','2026-01-01');"
request POST '/api/work-orders' "{\"PropertyID\":\"$p\",\"Description\":\"reported issue\"}";wid=$(json_string ID "$HTTP_BODY");[[ $HTTP_STATUS == 201 && -n $wid ]]||fail "work order setup failed status=$HTTP_STATUS body=$HTTP_BODY"
request POST "/api/work-orders/$wid/confirm" '{}';confirm=$HTTP_STATUS
request GET "/api/work-orders/$wid";detail=$HTTP_BODY
state=$(sql "SELECT CONCAT(status,'|',tenant_confirmed,'|',COALESCE(assignee_id,'')) FROM work_orders WHERE id='$wid';");history=$(sql "SELECT COUNT(*) FROM work_order_state_history WHERE work_order_id='$wid';")
[[ $confirm == 409 ]]||fail "early confirmation expected HTTP 409 got $confirm"
[[ $state == 'reported|0|' ]]||fail "reported work order was changed early got $state"
[[ $history == 1 ]]||fail "rejected confirmation changed history count=$history"
[[ $detail == *'"Status":"reported"'* ]]||fail "API detail status mismatch body=$detail"
finish 5
