#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-020;source scripts/verify/testlib.sh;login_admin
prefix=000-BUG020-$(new_id);ids=();for i in 1 2 3 4 5 6;do id=$(new_id);ids+=("$id");room=$(printf '%02d' "$i");sql_exec "INSERT INTO properties(id,building,room,status,available_from) VALUES('$id','$prefix','$room','maintenance','2026-01-01');";done
request GET '/api/properties?page=2&page_size=2&status=maintenance';body=$HTTP_BODY
[[ $HTTP_STATUS == 200 ]]||fail "property list expected HTTP 200 got $HTTP_STATUS"
[[ $body == *'"page":2'* && $body == *'"page_size":2'* ]]||fail "pagination metadata mismatch body=$body"
[[ $body == *"${ids[2]}"*"${ids[3]}"* ]]||fail "page 2 expected ids ${ids[2]} then ${ids[3]} body=$body"
[[ $body != *"${ids[0]}"* && $body != *"${ids[1]}"* && $body != *"${ids[4]}"* && $body != *"${ids[5]}"* ]]||fail 'page leaked rows outside expected fixture slice'
finish 4
