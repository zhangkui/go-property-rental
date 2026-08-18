#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-018;source scripts/verify/testlib.sh;login_admin
t=$(new_id);sql_exec "INSERT INTO tenants(id,name,phone,email,identity_no,status) VALUES('$t','Old Name','13000000000','old@example.com','OLD-ID','active');"
request PUT "/api/tenants/$t" '{"Name":"Updated Tenant","Phone":"13912345678","Email":"updated@example.com","IdentityNo":"ID-2026-018"}';update=$HTTP_STATUS
request GET "/api/tenants/$t";detail=$HTTP_BODY;state=$(sql "SELECT CONCAT(name,'|',phone,'|',email,'|',identity_no,'|',status) FROM tenants WHERE id='$t';")
[[ $update == 200 ]]||fail "tenant update expected HTTP 200 got $update"
[[ $detail == *'"Name":"Updated Tenant"'* ]]||fail "detail name mismatch body=$detail"
[[ $detail == *'"Phone":"13912345678"'* ]]||fail "detail phone mismatch body=$detail"
[[ $detail == *'"Email":"updated@example.com"'* ]]||fail "detail email mismatch body=$detail"
[[ $detail == *'"IdentityNo":"ID-2026-018"'* ]]||fail "detail identity number mismatch body=$detail"
[[ $state == 'Updated Tenant|13912345678|updated@example.com|ID-2026-018|active' ]]||fail "persisted tenant fields mismatch got $state"
finish 6
