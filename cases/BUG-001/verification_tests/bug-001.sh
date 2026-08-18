#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
BUG_ID=BUG-001
source scripts/verify/testlib.sh
login_admin
property_id=$(new_id)
room=${property_id:0:8}
sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$property_id','BUG001','$room','occupied','2026-01-01');"
request PATCH "/api/properties/$property_id/status" '{"status":"retired"}'
saved_status=$(sql "SELECT status FROM properties WHERE id='$property_id';")
audit_count=$(sql "SELECT COUNT(*) FROM audit_logs WHERE resource='property' AND resource_id='$property_id' AND action='property.status_changed';")
[[ $HTTP_STATUS == 400 ]] || fail "invalid property status expected HTTP 400 got $HTTP_STATUS body=$HTTP_BODY"
[[ $saved_status == occupied ]] || fail "rejected status must leave property occupied got $saved_status"
[[ $audit_count == 0 ]] || fail "rejected status must not create success audit got $audit_count"
finish 3
