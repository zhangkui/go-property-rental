#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-014;source scripts/verify/testlib.sh;login_admin
p=$(new_id);t=$(new_id);l=$(new_id);seed=$(new_id);room=${p:0:8}
sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$p','BUG014','$room','occupied','2026-01-01');INSERT INTO tenants(id,name,phone,email,identity_no,status) VALUES('$t','Settlement Tenant','','','','active');INSERT INTO leases(id,property_id,tenant_id,status,start_date,end_date,monthly_rent,deposit) VALUES('$l','$p','$t','terminating','2026-01-01','2026-12-31',10000,5000);INSERT INTO deposit_ledgers(id,lease_id,kind,reference,amount,reason,actor_id) VALUES('$seed','$l','collect','BUG014-SEED-$l',5000,'initial deposit',NULL);"
body="{\"lease_id\":\"$l\",\"deposit_deduction\":2000,\"items\":[{\"kind\":\"damage\",\"description\":\"wall repair\",\"amount\":2000},{\"kind\":\"utility\",\"description\":\"final utility\",\"amount\":800}],\"readings\":[{\"kind\":\"electricity\",\"reading\":3210},{\"kind\":\"water\",\"reading\":876}]}"
request POST '/api/settlements' "$body";created=$HTTP_BODY;sid=$(json_string ID "$created")
state=$(sql "SELECT CONCAT(status,'|',total,'|',deposit_deduction,'|',refund) FROM settlements WHERE id='$sid';");ledger=$(sql "SELECT GROUP_CONCAT(CONCAT(kind,':',amount) ORDER BY kind SEPARATOR '|') FROM deposit_ledgers WHERE lease_id='$l' AND reference LIKE 'settlement-%';");available=$(sql "SELECT SUM(CASE WHEN kind IN ('collect','topup') THEN amount WHEN kind IN ('deduct','refund') THEN -amount ELSE 0 END) FROM deposit_ledgers WHERE lease_id='$l';");items=$(sql "SELECT COUNT(*) FROM settlement_items WHERE settlement_id='$sid';");readings=$(sql "SELECT GROUP_CONCAT(CONCAT(kind,':',reading) ORDER BY kind SEPARATOR '|') FROM meter_readings WHERE lease_id='$l';")
[[ $HTTP_STATUS == 201 ]]||fail "settlement creation expected HTTP 201 got $HTTP_STATUS body=$created"
[[ $state == 'draft|2800|2000|3000' ]]||fail "settlement money mismatch got $state"
[[ $ledger == 'deduct:2000|refund:3000' ]]||fail "settlement deposit ledger mismatch got $ledger"
[[ $available == 0 ]]||fail "settlement must consume/refund full deposit got available=$available"
[[ $items == 2 ]]||fail "expected two settlement items got $items"
[[ $readings == 'electricity:3210|water:876' ]]||fail "meter readings mismatch got $readings"
[[ $created == *'2000'* && $created == *'3000'* ]]||fail "API settlement deduction/refund mismatch body=$created"
finish 7
