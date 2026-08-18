#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-027;source scripts/verify/testlib.sh;login_admin
uid=$(sql "SELECT id FROM users WHERE username='admin';");p=$(new_id);room=${p:0:8};sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG027','$room','available','2026-01-01');"
redis_key="dashboard:summary:$uid";redis_cmd DEL "$redis_key" >/dev/null
request GET '/api/dashboard';before=$HTTP_BODY
[[ $before == *'"Properties"'* && $before == *'"Total":1'* ]]||fail "dashboard precondition missing property totals body=$before"
redis_cmd SET "$redis_key" 'not-json' >/dev/null
request GET '/api/dashboard';after=$HTTP_BODY
[[ $HTTP_STATUS == 200 ]]||fail "dashboard fallback expected HTTP 200 got $HTTP_STATUS"
[[ $after == *'"GeneratedAt"'* && $after != *'0001-01-01'* ]]||fail "fallback response lost generated_at body=$after"
[[ $after == *'"Properties"'* && $after == *'"Total":1'* && $after == *'"Available":1'* ]]||fail "fallback response lost property counts body=$after"
finish 5
