#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-021;source scripts/verify/testlib.sh
suffix=${RANDOM}$(date +%s);user=bug021_$suffix;password='User123!Pass';TOKEN=''
request POST '/api/auth/register' "{\"username\":\"$user\",\"password\":\"$password\",\"display_name\":\"Disabled session user\",\"email\":\"$suffix@example.test\"}";uid=$(sql "SELECT id FROM users WHERE username='$user';")
request POST '/api/auth/login' "{\"username\":\"$user\",\"password\":\"$password\"}";user_access=$(json_string access_token "$HTTP_BODY");user_refresh=$(json_string refresh_token "$HTTP_BODY");TOKEN=$user_access;request GET '/api/me';before=$HTTP_STATUS
TOKEN='';login_admin;request PATCH "/api/users/$uid/status" '{"status":"disabled"}';disable=$HTTP_STATUS
TOKEN=$user_access;request GET '/api/me';access=$HTTP_STATUS
TOKEN='';request POST '/api/auth/refresh' "{\"refresh_token\":\"$user_refresh\"}";refresh=$HTTP_STATUS
request POST '/api/auth/login' "{\"username\":\"$user\",\"password\":\"$password\"}";new_login=$HTTP_STATUS
db=$(sql "SELECT status FROM users WHERE id='$uid';")
[[ $before == 200 ]]||fail "active user token precondition expected 200 got $before"
[[ $disable == 200 && $db == disabled ]]||fail "user disable precondition failed HTTP=$disable db=$db"
[[ $access == 401 ]]||fail "disabled user access token expected 401 got $access"
[[ $refresh == 401 ]]||fail "disabled user refresh token expected 401 got $refresh"
[[ $new_login == 401 ]]||fail "disabled user new login expected 401 got $new_login"
finish 5
