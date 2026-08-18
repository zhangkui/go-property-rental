#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-008;source scripts/verify/testlib.sh;login_admin
b=$(new_id);l=$(new_id)
sql_exec "INSERT INTO bills(id,lease_id,period_start,period_end,due_date,status,amount,penalty,discount,paid) VALUES('$b','$l','2026-07-01','2026-08-01','2026-08-01','paid',10000,0,0,10000);"
request POST "/api/bills/$b/late-fee" '{"at":"2026-08-18","daily_basis_points":100,"max_days":30}'
request GET "/api/bills/$b";detail_status=$HTTP_STATUS;detail=$HTTP_BODY
state=$(sql "SELECT CONCAT(status,'|',paid,'|',penalty,'|',amount+penalty-discount-paid) FROM bills WHERE id='$b';");adjustments=$(sql "SELECT COUNT(*) FROM bill_adjustments WHERE bill_id='$b';")
[[ $detail_status == 200 ]]||fail "bill detail expected HTTP 200 got $detail_status"
[[ $state == 'paid|10000|0|0' ]]||fail "paid bill state changed or regained balance got $state"
[[ $adjustments == 0 ]]||fail "adjustment ledger written for closed bill count=$adjustments"
[[ $detail == *'"status":"paid"'* || $detail == *'"Status":"paid"'* ]]||fail "API detail no longer reports paid body=$detail"
[[ $detail == *'"penalty":0'* || $detail == *'"Penalty":0'* ]]||fail "API detail reports late penalty body=$detail"
finish 5
