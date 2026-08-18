#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-029;source scripts/verify/testlib.sh;login_admin
admin=$(sql "SELECT id FROM users WHERE username='admin';");uid=$(new_id);name=bug029-${uid:0:8};sql_exec "INSERT INTO users(id,username,password_hash,display_name,email,status) SELECT '$uid','$name',password_hash,'Other Reviewer','','active' FROM users WHERE username='admin';INSERT INTO user_roles(user_id,role_id) SELECT '$uid',id FROM roles WHERE name='business_admin';"
TOKEN='';request POST '/api/auth/login' "{\"username\":\"$name\",\"password\":\"Admin123!\"}";other_token=$(json_string access_token "$HTTP_BODY")
TOKEN='';login_admin
body1="{\"request_type\":\"deposit_refund\",\"resource_type\":\"lease\",\"resource_id\":\"$(new_id)\",\"title\":\"BUG029 approval\",\"summary\":\"review order\",\"reviewers\":[\"$admin\",\"$uid\"]}"
request POST '/api/approvals' "$body1";a1=$(json_string ID "$HTTP_BODY");body2="{\"request_type\":\"deposit_refund\",\"resource_type\":\"lease\",\"resource_id\":\"$(new_id)\",\"title\":\"BUG029 approval\",\"summary\":\"review order\",\"reviewers\":[\"$admin\",\"$uid\"]}";request POST '/api/approvals' "$body2";a2=$(json_string ID "$HTTP_BODY")
TOKEN=$other_token;request POST "/api/approvals/$a1/decision" '{"decision":"approved","comment":"checked"}';unauth=$HTTP_STATUS
TOKEN='';login_admin;request POST "/api/approvals/$a2/decision" '{"decision":"approved","comment":"checked"}';auth=$HTTP_STATUS
state1=$(sql "SELECT CONCAT(status,'|',current_reviewer_id,'|',(SELECT decision FROM approval_steps WHERE approval_id='$a1' ORDER BY sequence_no LIMIT 1)) FROM approval_requests WHERE id='$a1';");state2=$(sql "SELECT CONCAT(status,'|',current_reviewer_id,'|',(SELECT decision FROM approval_steps WHERE approval_id='$a2' ORDER BY sequence_no LIMIT 1)) FROM approval_requests WHERE id='$a2';");actor=$(sql "SELECT COALESCE(actor_id,'') FROM approval_state_history WHERE approval_id='$a2' AND comment='checked' ORDER BY created_at DESC LIMIT 1;")
[[ $unauth == 400 ]]||fail "non-current reviewer expected HTTP 400 got $unauth"
[[ $state1 == "pending|$admin|pending" ]]||fail "unauthorized decision changed approval got $state1"
[[ $auth == 200 ]]||fail "current reviewer expected HTTP 200 got $auth"
[[ $state2 == "pending|$uid|approved" ]]||fail "first approval/current reviewer mismatch got $state2"
[[ $actor == "$admin" ]]||fail "decision history actor expected admin got $actor"
finish 5
