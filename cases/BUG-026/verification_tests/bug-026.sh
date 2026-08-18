#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-026;source scripts/verify/testlib.sh;login_admin
uid=$(new_id);p=$(new_id);t=$(new_id);l=$(new_id);room=${p:0:8};username=bug026-$room;today=$(date -u +%F)
sql_exec "INSERT INTO users(id,username,password_hash,display_name,email,status) SELECT '$uid','$username',password_hash,'Reminder Operator','','active' FROM users WHERE username='admin';INSERT INTO user_roles(user_id,role_id) SELECT '$uid',id FROM roles WHERE name='daily_operator';INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG026','$room','occupied','2026-01-01');INSERT INTO tenants(id,name,phone,email,identity_no,status) VALUES('$t','Reminder Tenant','','','','active');INSERT INTO leases(id,property_id,tenant_id,status,start_date,end_date,monthly_rent,deposit) VALUES('$l','$p','$t','active','2026-01-01','2026-09-01',10000,0);"
request POST '/api/reminders/run';one=$HTTP_STATUS;request POST '/api/reminders/run';two=$HTTP_STATUS
rule=$(sql "SELECT id FROM reminder_rules WHERE category='lease_expiry' LIMIT 1;");runs=$(sql "SELECT COUNT(*) FROM reminder_runs WHERE rule_id='$rule' AND run_key LIKE '$today%';");notes=$(sql "SELECT COUNT(*) FROM notifications WHERE user_id='$uid' AND resource_type='lease' AND resource_id='$l' AND category='lease_expiry';")
[[ $one == 200 && $two == 200 ]]||fail "repeated reminder runs expected HTTP 200/200 got $one/$two"
[[ $runs == 1 ]]||fail "daily reminder rule expected one run row got $runs"
[[ $notes == 1 ]]||fail "recipient/lease expected one notification got $notes"
finish 3
