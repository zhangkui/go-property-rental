#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-022;source scripts/verify/testlib.sh;login_admin
admin_token=$TOKEN;uid=$(new_id);name=bug022-${uid:0:8};sql_exec "INSERT INTO users(id,username,password_hash,display_name,email,status) SELECT '$uid','$name',password_hash,'BUG022 User','','active' FROM users WHERE username='admin';INSERT INTO user_roles(user_id,role_id) SELECT '$uid',id FROM roles WHERE name='business_admin';"
TOKEN='';request POST '/api/auth/login' "{\"username\":\"$name\",\"password\":\"Admin123!\"}";old_token=$(json_string access_token "$HTTP_BODY")
readonly=$(sql "SELECT id FROM roles WHERE name='readonly_auditor';");TOKEN=$admin_token;request PUT "/api/users/$uid/roles" "{\"role_ids\":[\"$readonly\"]}";replace=$HTTP_STATUS
TOKEN=$old_token;request POST '/api/properties' "{\"building\":\"BUG022\",\"room\":\"${uid:0:8}\",\"status\":\"available\"}";old_status=$HTTP_STATUS
TOKEN='';request POST '/api/auth/login' "{\"username\":\"$name\",\"password\":\"Admin123!\"}";TOKEN=$(json_string access_token "$HTTP_BODY");request GET '/api/me';me=$HTTP_BODY
roles=$(sql "SELECT GROUP_CONCAT(r.name ORDER BY r.name) FROM user_roles ur JOIN roles r ON r.id=ur.role_id WHERE ur.user_id='$uid';")
[[ $replace == 200 ]]||fail "role replacement expected HTTP 200 got $replace"
[[ $old_status == 401 ]]||fail "old access token expected 401 got $old_status"
[[ $roles == readonly_auditor ]]||fail "database roles expected readonly_auditor got $roles"
[[ $me != *'property:create'* ]]||fail 'new session retained property:create permission'
finish 4
