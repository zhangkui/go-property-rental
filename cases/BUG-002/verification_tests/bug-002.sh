#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-002;source scripts/verify/testlib.sh;login_admin
p=$(new_id);a=$(new_id);b=$(new_id);c=$(new_id);x=${p:0:8}
sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG002','$x','available','2026-09-01');INSERT INTO facilities(id,name) VALUES('$a','A-$x'),('$b','B-$x'),('$c','C-$x');"
request PUT "/api/properties/$p/facilities" "{\"facility_ids\":[\"$a\",\"$b\"]}";[[ $HTTP_STATUS == 200 ]]||fail "initial replace expected HTTP 200 got $HTTP_STATUS"
request GET "/api/properties/$p/facilities";first=$HTTP_BODY
request PUT "/api/properties/$p/facilities" "{\"facility_ids\":[\"$c\"]}";[[ $HTTP_STATUS == 200 ]]||fail "second replace expected HTTP 200 got $HTTP_STATUS"
request GET "/api/properties/$p/facilities";second=$HTTP_BODY;count=$(sql "SELECT COUNT(*) FROM property_facilities WHERE property_id='$p';")
[[ $first == *"$a"* && $first == *"$b"* ]]||fail 'initial API detail did not expose both facilities'
[[ $second == *"$c"* ]]||fail 'requested replacement facility is missing from API detail'
[[ $second != *"$a"* && $second != *"$b"* ]]||fail "API detail retained stale facilities body=$second"
[[ $count == 1 ]]||fail "database retained stale facility associations count=$count"
finish 6
