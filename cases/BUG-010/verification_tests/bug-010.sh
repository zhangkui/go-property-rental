#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-010;source scripts/verify/testlib.sh;login_admin
b=$(new_id);l=$(new_id);ref=BUG010-$(new_id)
sql_exec "INSERT INTO bills(id,lease_id,period_start,period_end,due_date,status,amount,penalty,discount,paid) VALUES('$b','$l','2026-08-01','2026-09-01','2026-08-05','unpaid',10000,0,0,0);"
body="{\"reference\":\"$ref\",\"payer\":\"tenant\",\"amount\":5000,\"allocations\":[{\"bill_id\":\"$b\",\"amount\":5000}]}"
request POST '/api/payments' "$body";first=$HTTP_STATUS
request POST '/api/payments' "$body";second=$HTTP_STATUS
request GET "/api/bills/$b";detail=$HTTP_BODY
receipts=$(sql "SELECT COUNT(*) FROM payment_receipts WHERE reference LIKE '$ref%';");allocations=$(sql "SELECT COUNT(*) FROM payment_allocations WHERE bill_id='$b';");stored=$(sql "SELECT reference FROM payment_receipts WHERE reference LIKE '$ref%' ORDER BY paid_at LIMIT 1;");state=$(sql "SELECT CONCAT(status,'|',paid,'|',amount+penalty-discount-paid) FROM bills WHERE id='$b';")
[[ $first == 201 ]]||fail "first payment expected HTTP 201 got $first"
[[ $second != 201 ]]||fail 'duplicate payment reference was accepted'
[[ $receipts == 1 ]]||fail "expected one payment receipt got $receipts"
[[ $allocations == 1 ]]||fail "expected one allocation got $allocations"
[[ $state == 'partial|5000|5000' ]]||fail "bill expected partial paid/outstanding 5000/5000 got $state"
[[ $detail == *'5000'* ]]||fail "bill API detail missing paid amount body=$detail"
[[ $stored == "$ref" ]]||fail "stored reference differs from client key got $stored"
finish 7
