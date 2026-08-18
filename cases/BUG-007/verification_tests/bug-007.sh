#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-007;source scripts/verify/testlib.sh;login_admin
p=$(new_id);t=$(new_id);l=$(new_id);key=BUG007-$(new_id);room=${p:0:8}
sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG007','$room','occupied','2026-01-01');INSERT INTO tenants(id,name,phone,email,identity_no,status) VALUES('$t','Billing Tenant','','','','active');INSERT INTO leases(id,property_id,tenant_id,status,start_date,end_date,monthly_rent,deposit) VALUES('$l','$p','$t','active','2026-01-01','2027-12-31',88000,30000);"
body="{\"lease_id\":\"$l\",\"period_start\":\"2026-09-01\",\"idempotency_key\":\"$key\"}"
request POST '/api/bills/generate' "$body";first_status=$HTTP_STATUS;first_body=$HTTP_BODY;bill_id=$(sql "SELECT id FROM bills WHERE lease_id='$l' AND period_start='2026-09-01' ORDER BY id LIMIT 1;")
request POST '/api/bills/generate' "$body";second_status=$HTTP_STATUS;second_body=$HTTP_BODY
bill_count=$(sql "SELECT COUNT(*) FROM bills WHERE lease_id='$l' AND period_start='2026-09-01';");key_count=$(sql "SELECT COUNT(*) FROM idempotency_keys WHERE scope='bill.generate' AND request_key='$key' AND resource_id='$bill_id';")
[[ $first_status == 201 && $first_body == *'"created":true'* ]]||fail "first generation expected HTTP 201 created=true got $first_status body=$first_body"
[[ $second_status == 200 && $second_body == *'"created":false'* ]]||fail "duplicate generation expected HTTP 200 created=false got $second_status body=$second_body"
[[ -n $bill_id && $first_body == *"$bill_id"* && $second_body == *"$bill_id"* ]]||fail 'duplicate request did not return the same bill id'
[[ $bill_count == 1 ]]||fail "duplicate request created $bill_count bills"
[[ $key_count == 1 ]]||fail "expected one persisted idempotency key got $key_count"
finish 5
