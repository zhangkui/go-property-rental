#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-009;source scripts/verify/testlib.sh;login_admin
admin=$(sql "SELECT id FROM users WHERE username='admin';");other=$(new_id);name=bug009-${other:0:8};own=$(new_id);control=$(new_id);foreign=$(new_id)
sql_exec "INSERT INTO users(id,username,password_hash,display_name,email,status) SELECT '$other','$name',password_hash,'Other User','','active' FROM users WHERE username='admin';INSERT INTO notifications(id,user_id,category,title,content,status) VALUES('$own','$admin','lease','Own Target','target','unread'),('$control','$admin','lease','Own Control','control','unread'),('$foreign','$other','lease','Foreign','foreign','unread');"
request POST "/api/notifications/$own/read" '{}';own_http=$HTTP_STATUS
request POST "/api/notifications/$foreign/read" '{}';foreign_http=$HTTP_STATUS
own_state=$(sql "SELECT CONCAT(status,'|',read_at IS NOT NULL) FROM notifications WHERE id='$own';");control_state=$(sql "SELECT status FROM notifications WHERE id='$control';");foreign_state=$(sql "SELECT CONCAT(status,'|',read_at IS NULL) FROM notifications WHERE id='$foreign';")
request GET '/api/notifications?status=unread&page=1&page_size=20';inbox=$HTTP_BODY;unread=$(json_number unread "$inbox")
[[ $own_http == 200 ]]||fail "own notification mark-read expected HTTP 200 got $own_http"
[[ $foreign_http == 200 ]]||fail "foreign notification no-op expected stable HTTP 200 got $foreign_http"
[[ $own_state == 'read|1' ]]||fail "own target notification expected read with timestamp got $own_state"
[[ $control_state == unread ]]||fail "unrelated own notification must remain unread got $control_state"
[[ $foreign_state == 'unread|1' ]]||fail "foreign notification must remain unread without timestamp got $foreign_state"
[[ $unread == 1 ]]||fail "admin unread count expected one control notification got $unread body=$inbox"
finish 6
