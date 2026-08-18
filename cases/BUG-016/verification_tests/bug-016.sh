#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-016;source scripts/verify/testlib.sh;login_admin
p=$(new_id);assignee=$(new_id);room=${p:0:8};sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG016','$room','available','2026-01-01');"
request POST '/api/work-orders' "{\"PropertyID\":\"$p\",\"Description\":\"material calculation\"}";wid=$(json_string ID "$HTTP_BODY")
request POST "/api/work-orders/$wid/assign" "{\"assignee_id\":\"$assignee\"}";[[ $HTTP_STATUS == 200 ]]||fail "assignment setup failed status=$HTTP_STATUS"
request POST "/api/work-orders/$wid/materials" '{"name":"filter","quantity":3,"unit_cost":2500}';[[ $HTTP_STATUS == 201 ]]||fail "material add expected HTTP 201 got $HTTP_STATUS"
request GET "/api/work-orders/$wid";detail=$HTTP_BODY
state=$(sql "SELECT CONCAT(status,'|',material_cost) FROM work_orders WHERE id='$wid';");row=$(sql "SELECT CONCAT(COUNT(*),'|',MIN(quantity),'|',MIN(unit_cost),'|',SUM(quantity*unit_cost)) FROM work_order_materials WHERE work_order_id='$wid';")
[[ $state == 'assigned|7500' ]]||fail "work order status/material cost expected assigned/7500 got $state"
[[ $row == '1|3|2500|7500' ]]||fail "material row quantity/cost calculation mismatch got $row"
[[ $detail == *'"Quantity":3'* ]]||fail "API material quantity mismatch body=$detail"
[[ $detail == *'"UnitCost":2500'* && $detail == *'"MaterialCost":7500'* ]]||fail "API material cost mismatch body=$detail"
finish 6
