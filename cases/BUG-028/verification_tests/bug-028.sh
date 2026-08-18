#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-028;source scripts/verify/testlib.sh;login_admin
p=$(new_id);t=$(new_id);l=$(new_id);room=${p:0:8};sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG028','$room','occupied','2026-01-01');INSERT INTO tenants(id,name,phone,email,identity_no,status) VALUES('$t','Report Boundary Tenant','','','','active');INSERT INTO leases(id,property_id,tenant_id,status,start_date,end_date,monthly_rent,deposit) VALUES('$l','$p','$t','active','2026-01-01','2026-08-18',12345,6789);"
request GET "/api/reports/rent-roll?property_id=$p&start_date=2026-08-18&end_date=2026-08-18";api_body=$HTTP_BODY
request GET "/api/reports/rent-roll/export?property_id=$p&start_date=2026-08-18&end_date=2026-08-18";csv=$HTTP_BODY
[[ $api_body == *"$l"* ]]||fail "boundary rent-roll API omitted lease $l body=$api_body"
[[ $api_body == *'12345'* ]]||fail 'boundary rent-roll API omitted rent amount'
[[ $csv == *'Report Boundary Tenant'* ]]||fail 'CSV export omitted boundary lease tenant'
[[ $csv == *'12345'* ]]||fail 'CSV export omitted boundary lease rent'
finish 4
