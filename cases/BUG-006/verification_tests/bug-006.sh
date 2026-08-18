#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-006;source scripts/verify/testlib.sh;login_admin
p=$(new_id);t=$(new_id);l=$(new_id);v1=$(new_id);v2=$(new_id);room=${p:0:8}
sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG006','$room','occupied','2026-01-01');INSERT INTO tenants(id,name,phone,email,identity_no,status) VALUES('$t','Version Tenant','','','','active');INSERT INTO leases(id,property_id,tenant_id,status,start_date,end_date,monthly_rent,deposit) VALUES('$l','$p','$t','active','2026-01-01','2027-12-31',250000,300000);INSERT INTO lease_versions(id,lease_id,version_no,monthly_rent,deposit,start_date,end_date) VALUES('$v1','$l',1,200000,200000,'2026-01-01','2026-12-31'),('$v2','$l',2,250000,300000,'2027-01-01','2027-12-31');"
request GET "/api/leases/$l";body=$HTTP_BODY;db=$(sql "SELECT CONCAT(COUNT(*),'|',MAX(version_no),'|',MAX(monthly_rent),'|',MAX(deposit)) FROM lease_versions WHERE lease_id='$l';")
[[ $HTTP_STATUS == 200 ]]||fail "lease detail expected HTTP 200 got $HTTP_STATUS"
[[ $body == *'"VersionNo":2'*'"VersionNo":1'* ]]||fail "latest version must be first body=$body"
[[ $body == *'"VersionNo":2,"MonthlyRent":250000,"Deposit":300000'* ]]||fail 'latest version money mismatch'
[[ $body == *'"VersionNo":1,"MonthlyRent":200000,"Deposit":200000'* ]]||fail 'historical version payload mismatch'
[[ $db == '2|2|250000|300000' ]]||fail "database version precondition mismatch got $db"
[[ $body == *'"MonthlyRent":250000,"Deposit":300000'* ]]||fail 'lease current money does not match latest version'
finish 6
