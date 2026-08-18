#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-023;source scripts/verify/testlib.sh
suffix=${RANDOM}$(date +%s);user=bug023-$suffix;uid=$(new_id);sql_exec "INSERT INTO users(id,username,password_hash,display_name,email,status) SELECT '$uid','$user',password_hash,'BUG023','','active' FROM users WHERE username='admin';";redis_cmd DEL "auth:attempt:$user" "auth:attempt:$user:10.23.0.1" "auth:attempt:$user:10.23.0.2" >/dev/null
login_ip(){ local ip=$1 pass=$2 r;r=$(curl -sS -X POST "$API_URL/api/auth/login" -H "X-Forwarded-For: $ip" -H 'Content-Type: application/json' --data "{\"username\":\"$user\",\"password\":\"$pass\"}" -w $'\n%{http_code}');LS=${r##*$'\n'};LB=${r%$'\n'*};}
for attempt in 1 2 3 4;do login_ip 10.23.0.1 'Wrong123!';[[ $LS == 401 && $LB == *'invalid credentials'* ]]||fail "pre-success failure $attempt must remain below limit status=$LS body=$LB";done
login_ip 10.23.0.1 'Admin123!';success=$LS;success_body=$LB
for attempt in 1 2 3 4 5;do login_ip 10.23.0.1 'Wrong123!';[[ $LS == 401 && $LB == *'invalid credentials'* ]]||fail "post-reset failure $attempt must remain allowed through boundary 5 status=$LS body=$LB";done
login_ip 10.23.0.1 'Wrong123!';limited_status=$LS;limited_body=$LB
login_ip 10.23.0.2 'Wrong123!';other_status=$LS;other_body=$LB
limited=$(sql "SELECT COUNT(*) FROM audit_logs WHERE action='auth.login.rate_limited' AND JSON_UNQUOTE(JSON_EXTRACT(detail,'$.username'))='$user';")
[[ $success == 200 ]]||fail "valid login after four failures expected 200 got $success body=$success_body"
[[ $limited_status == 401 && $limited_body == *'too many login attempts'* ]]||fail "sixth post-reset failure must be rate limited status=$limited_status body=$limited_body"
[[ $other_status == 401 && $other_body == *'invalid credentials'* ]]||fail "independent IP first failure must not be rate limited status=$other_status body=$other_body"
[[ ${limited:-0} -ge 1 ]]||fail 'fixture did not exercise rate-limited path'
finish 16
