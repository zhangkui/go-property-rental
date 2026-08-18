#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-012;source scripts/verify/testlib.sh;login_admin
l=$(new_id);ref=BUG012-$(new_id);body="{\"kind\":\"collect\",\"reference\":\"$ref\",\"reason\":\"initial deposit\",\"amount\":5000}"
request POST "/api/leases/$l/deposits" "$body";first=$HTTP_STATUS
request POST "/api/leases/$l/deposits" "$body";second=$HTTP_STATUS
request GET "/api/leases/$l/deposits";detail=$HTTP_BODY
count=$(sql "SELECT COUNT(*) FROM deposit_ledgers WHERE lease_id='$l';");sum=$(sql "SELECT COALESCE(SUM(amount),0) FROM deposit_ledgers WHERE lease_id='$l' AND kind='collect';");stored=$(sql "SELECT reference FROM deposit_ledgers WHERE lease_id='$l' ORDER BY created_at LIMIT 1;")
[[ $first == 201 ]]||fail "first deposit collection expected HTTP 201 got $first"
[[ $second != 201 ]]||fail 'duplicate deposit reference was accepted'
[[ $count == 1 ]]||fail "expected one ledger row got $count"
[[ $sum == 5000 ]]||fail "received and available balance expected 5000 got $sum"
[[ $detail == *'5000'* ]]||fail "deposit API balance missing 5000 body=$detail"
[[ $stored == "$ref" ]]||fail "stored reference differs from client key got $stored"
finish 6
