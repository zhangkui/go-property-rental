#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-024;source scripts/verify/testlib.sh;TOKEN=''
request POST '/api/auth/login' '{"username":"admin","password":"Admin123!"}';login_http=$HTTP_STATUS;TOKEN=$(json_string access_token "$HTTP_BODY");initial_token=$TOKEN;uid=$(sql "SELECT id FROM users WHERE username='admin';");before_hash=$(sql "SELECT password_hash FROM users WHERE id='$uid';")
request POST '/api/me/password' '{"current_password":"Admin123!","new_password":"NewAdmin456!"}';change_http=$HTTP_STATUS;change_body=$HTTP_BODY;after_hash=$(sql "SELECT password_hash FROM users WHERE id='$uid';");active_after_change=$(sql "SELECT COUNT(*) FROM auth_sessions WHERE user_id='$uid' AND revoked_at IS NULL;");audit=$(sql "SELECT COUNT(*) FROM audit_logs WHERE action='auth.password.changed' AND resource_id='$uid';")
request GET '/api/me';old_token_http=$HTTP_STATUS
TOKEN='';request POST '/api/auth/login' '{"username":"admin","password":"Admin123!"}';old_password_http=$HTTP_STATUS
request POST '/api/auth/login' '{"username":"admin","password":"NewAdmin456!"}';new_password_http=$HTTP_STATUS
[[ $login_http == 200 && -n $initial_token ]]||fail "initial admin login expected HTTP 200"
[[ $change_http == 200 ]]||fail "password change expected HTTP 200 got $change_http body=$change_body"
[[ $before_hash != "$after_hash" ]]||fail 'password hash must change after password update'
[[ $active_after_change == 0 ]]||fail "all previous sessions must be revoked got active=$active_after_change"
[[ $audit == 1 ]]||fail "password change expected one audit row got $audit"
[[ $old_token_http == 401 ]]||fail "old access token expected HTTP 401 after password change got $old_token_http"
[[ $old_password_http == 401 ]]||fail "old password expected HTTP 401 after change got $old_password_http"
[[ $new_password_http == 200 ]]||fail "new password expected HTTP 200 got $new_password_http"
finish 8
