#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-005;source scripts/verify/testlib.sh;login_admin
p=$(new_id);t=$(new_id);l=$(new_id);v=$(new_id);room=${p:0:8}
sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG005','$room','occupied','2026-01-01');INSERT INTO tenants(id,name,phone,email,identity_no,status) VALUES('$t','Renew Tenant','','','','active');INSERT INTO leases(id,property_id,tenant_id,status,start_date,end_date,monthly_rent,deposit) VALUES('$l','$p','$t','active','2026-01-01','2026-12-31',100000,40000);INSERT INTO lease_versions(id,lease_id,version_no,monthly_rent,deposit,start_date,end_date) VALUES('$v','$l',1,100000,40000,'2026-01-01','2026-12-31');"
request POST "/api/leases/$l/renew" '{"end_date":"2027-12-31","monthly_rent":120000,"deposit":50000}'
lease=$(sql "SELECT CONCAT(monthly_rent,'|',deposit,'|',end_date) FROM leases WHERE id='$l';");version=$(sql "SELECT CONCAT(version_no,'|',monthly_rent,'|',deposit,'|',start_date,'|',end_date) FROM lease_versions WHERE lease_id='$l' ORDER BY version_no DESC LIMIT 1;");history=$(sql "SELECT COUNT(*) FROM lease_state_history WHERE lease_id='$l' AND reason='renewed';")
[[ $HTTP_STATUS == 200 ]]||fail "renewal expected HTTP 200 got $HTTP_STATUS body=$HTTP_BODY"
[[ $lease == '120000|50000|2027-12-31' ]]||fail "lease renewal money/date mismatch got $lease"
[[ $version == '2|120000|50000|2026-12-31|2027-12-31' ]]||fail "contract version money/boundaries mismatch got $version"
[[ $history == 1 ]]||fail "successful renewal expected one history row got $history"
finish 4
