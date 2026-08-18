#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-017;source scripts/verify/testlib.sh;login_admin
admin=$(sql "SELECT id FROM users WHERE username='admin';");p=$(new_id);w=$(new_id);assignee=$(new_id);room=${p:0:8};username=bug017-$room
sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG017','$room','maintenance','2026-01-01');INSERT INTO users(id,username,password_hash,display_name,email,status) VALUES('$assignee','$username','not-used','Maintenance Tech','','active');INSERT INTO work_orders(id,property_id,tenant_id,assignee_id,description,status,material_cost,tenant_confirmed,created_at,updated_at) VALUES('$w','$p',NULL,NULL,'Broken faucet','reported',0,FALSE,UTC_TIMESTAMP(),UTC_TIMESTAMP());"
request POST "/api/work-orders/$w/assign" "{\"assignee_id\":\"$assignee\"}"
state=$(sql "SELECT CONCAT(status,'|',COALESCE(assignee_id,'')) FROM work_orders WHERE id='$w';");history=$(sql "SELECT CONCAT(from_status,'|',to_status,'|',reason,'|',COALESCE(actor_id,'')) FROM work_order_state_history WHERE work_order_id='$w' ORDER BY created_at DESC LIMIT 1;");audit=$(sql "SELECT CONCAT(COUNT(*),'|',COALESCE(MAX(actor_id),'')) FROM audit_logs WHERE action='work_order.action' AND resource='work_order' AND resource_id='$w';")
[[ $HTTP_STATUS == 200 ]]||fail "assignment expected HTTP 200 got $HTTP_STATUS body=$HTTP_BODY"
[[ $state == "assigned|$assignee" ]]||fail "work order must store requested assignee got $state"
[[ $history == "reported|assigned|assigned|$admin" ]]||fail "state history actor/transition mismatch got $history"
[[ $audit == "1|$admin" ]]||fail "assignment audit actor mismatch got $audit"
finish 4
