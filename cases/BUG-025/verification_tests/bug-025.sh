#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-025;source scripts/verify/testlib.sh;login_admin
actor=$(sql "SELECT id FROM users WHERE username='admin';");resource=bug025-$(new_id);i1=$(new_id);i2=$(new_id);decoy=$(new_id)
sql_exec "INSERT INTO audit_logs(id,actor_id,action,resource,resource_id,detail,created_at) VALUES('$i1','$actor','bug025.first','$resource','R1',JSON_OBJECT('n',1),'2026-08-18 01:00:00'),('$i2','$actor','bug025.second','$resource','R2',JSON_OBJECT('n',2),'2026-08-18 02:00:00'),('$decoy',NULL,'bug025.decoy','$resource','R3',JSON_OBJECT('n',3),'2026-08-18 03:00:00');"
request GET "/api/audit-logs?resource=$resource&actor_id=$actor&page=1&page_size=10";body=$HTTP_BODY
[[ $HTTP_STATUS == 200 ]]||fail "audit query expected HTTP 200 got $HTTP_STATUS"
[[ $body == *"$i2"*"$i1"* ]]||fail "expected newest-first ids $i2 then $i1 body=$body"
[[ $body != *"$decoy"* ]]||fail 'actor filter leaked decoy audit row'
[[ $body == *'bug025.second'* && $body == *'bug025.first'* ]]||fail 'resource filter omitted actor/resource audit rows'
finish 4
