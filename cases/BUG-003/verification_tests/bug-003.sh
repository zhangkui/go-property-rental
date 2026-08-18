#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-003;source scripts/verify/testlib.sh;login_admin
p=$(new_id);t=$(new_id);room=${p:0:8}
sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG003','$room','available','2026-09-01');INSERT INTO tenants(id,name,phone,email,identity_no,status) VALUES('$t','Disabled tenant','','','','disabled');"
request GET "/api/tenants/$t";raw_status=$(sql "SELECT status FROM tenants WHERE id='$t';");[[ $HTTP_STATUS == 200 && $raw_status == disabled ]]||fail "disabled tenant fixture mismatch HTTP=$HTTP_STATUS db=$raw_status body=$HTTP_BODY"
body="{\"property_id\":\"$p\",\"tenant_id\":\"$t\",\"start_date\":\"2026-09-01\",\"end_date\":\"2027-08-31\",\"monthly_rent\":200000,\"deposit\":200000,\"occupants\":[]}"
request POST '/api/leases' "$body"
leases=$(sql "SELECT COUNT(*) FROM leases WHERE tenant_id='$t';");versions=$(sql "SELECT COUNT(*) FROM lease_versions v JOIN leases l ON l.id=v.lease_id WHERE l.tenant_id='$t';");history=$(sql "SELECT COUNT(*) FROM lease_state_history h JOIN leases l ON l.id=h.lease_id WHERE l.tenant_id='$t';")
[[ $HTTP_STATUS == 409 ]]||fail "disabled tenant lease expected HTTP 409 got $HTTP_STATUS body=$HTTP_BODY"
[[ $leases == 0 ]]||fail "lease row created for disabled tenant count=$leases"
[[ $versions == 0 ]]||fail "lease version created for disabled tenant count=$versions"
[[ $history == 0 ]]||fail "lease history created for disabled tenant count=$history"
finish 5
