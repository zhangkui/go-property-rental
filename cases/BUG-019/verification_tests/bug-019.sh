#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-019;source scripts/verify/testlib.sh;login_admin
admin=$(sql "SELECT id FROM users WHERE username='admin';");p=$(new_id);t=$(new_id);l=$(new_id);s=$(new_id);room=${p:0:8}
sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG019','$room','occupied','2026-01-01');INSERT INTO tenants(id,name,phone,email,identity_no,status) VALUES('$t','Complete Tenant','','','','active');INSERT INTO leases(id,property_id,tenant_id,status,start_date,end_date,monthly_rent,deposit) VALUES('$l','$p','$t','terminating','2026-01-01','2026-12-31',10000,0);INSERT INTO settlements(id,lease_id,status,total,deposit_deduction,refund) VALUES('$s','$l','draft',0,0,0);"
request POST "/api/settlements/$s/complete"
state=$(sql "SELECT CONCAT(s.status,'|',l.status,'|',p.status,'|',IF(s.settled_at IS NULL,0,1)) FROM settlements s JOIN leases l ON l.id=s.lease_id JOIN properties p ON p.id=l.property_id WHERE s.id='$s';");history=$(sql "SELECT COALESCE(actor_id,'') FROM lease_state_history WHERE lease_id='$l' AND reason='move-out settlement' ORDER BY created_at DESC LIMIT 1;");audit=$(sql "SELECT COALESCE(MAX(actor_id),'') FROM audit_logs WHERE action='settlement.complete' AND resource_id='$s';")
[[ $HTTP_STATUS == 200 ]]||fail "completion expected HTTP 200 got $HTTP_STATUS body=$HTTP_BODY"
[[ $state == 'completed|closed|available|1' ]]||fail "settlement/lease/property atomic state mismatch got $state"
[[ $history == "$admin" ]]||fail "lease history actor expected admin got $history"
[[ $audit == "$admin" ]]||fail "completion audit actor expected admin got $audit"
finish 4
