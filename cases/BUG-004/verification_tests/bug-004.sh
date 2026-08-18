#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-004;source scripts/verify/testlib.sh;login_admin
admin=$(sql "SELECT id FROM users WHERE username='admin';");p=$(new_id);t=$(new_id);l=$(new_id);room=${p:0:8}
sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG004','$room','available','2026-01-01');INSERT INTO tenants(id,name,phone,email,identity_no,status) VALUES('$t','Transition Tenant','','','','active');INSERT INTO leases(id,property_id,tenant_id,status,start_date,end_date,monthly_rent,deposit) VALUES('$l','$p','$t','draft','2026-09-01','2027-08-31',100000,100000);INSERT INTO lease_state_history(id,lease_id,from_status,to_status,reason,actor_id) VALUES(UUID(),'$l','','draft','created','$admin');"
request POST "/api/leases/$l/transition" '{"status":"pending","reason":"submit for approval"}';response=$HTTP_BODY
state=$(sql "SELECT status FROM leases WHERE id='$l';");history=$(sql "SELECT CONCAT(from_status,'|',to_status,'|',reason,'|',actor_id) FROM lease_state_history WHERE lease_id='$l' AND reason='submit for approval' ORDER BY created_at DESC LIMIT 1;");count=$(sql "SELECT COUNT(*) FROM lease_state_history WHERE lease_id='$l';");audit=$(sql "SELECT COUNT(*) FROM audit_logs WHERE action='lease.transition' AND resource_id='$l';")
[[ $HTTP_STATUS == 200 ]]||fail "lease submit expected HTTP 200 got $HTTP_STATUS body=$response"
[[ $state == pending ]]||fail "lease expected pending after submit got $state"
[[ $history == "draft|pending|submit for approval|$admin" ]]||fail "lease history expected draft to pending with admin actor got $history"
[[ $count == 2 ]]||fail "lease expected created and submit history rows got $count"
[[ $audit == 1 ]]||fail "lease transition expected one audit row got $audit"
finish 5
