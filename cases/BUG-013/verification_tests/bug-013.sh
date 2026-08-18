#!/usr/bin/env bash
set -uo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../..";BUG_ID=BUG-013;source scripts/verify/testlib.sh;login_admin
admin=$(sql "SELECT id FROM users WHERE username='admin';");other=$(new_id);name=bug013-${other:0:8};own=$(new_id);foreign=$(new_id);reviewer=$(new_id)
sql_exec "INSERT INTO users(id,username,password_hash,display_name,email,status) SELECT '$other','$name',password_hash,'Other Applicant','','active' FROM users WHERE username='admin';INSERT INTO approval_requests(id,request_type,resource_type,resource_id,title,summary,status,applicant_id,current_reviewer_id) VALUES('$own','deposit_refund','lease','$(new_id)','Own approval','','pending','$admin','$reviewer'),('$foreign','deposit_refund','lease','$(new_id)','Foreign approval','','pending','$other','$reviewer');"
request POST "/api/approvals/$own/cancel" '{"comment":"withdraw own request"}';own_http=$HTTP_STATUS
request POST "/api/approvals/$foreign/cancel" '{"comment":"unauthorized withdrawal"}';foreign_http=$HTTP_STATUS
own_state=$(sql "SELECT CONCAT(status,'|',current_reviewer_id IS NULL,'|',completed_at IS NOT NULL) FROM approval_requests WHERE id='$own';");foreign_state=$(sql "SELECT CONCAT(status,'|',applicant_id) FROM approval_requests WHERE id='$foreign';");history_actor=$(sql "SELECT actor_id FROM approval_state_history WHERE approval_id='$own' AND to_status='cancelled' LIMIT 1;");audit_actor=$(sql "SELECT actor_id FROM audit_logs WHERE action='approval.cancel' AND resource_id='$own' LIMIT 1;");foreign_history=$(sql "SELECT COUNT(*) FROM approval_state_history WHERE approval_id='$foreign';");foreign_audit=$(sql "SELECT COUNT(*) FROM audit_logs WHERE action='approval.cancel' AND resource_id='$foreign';")
[[ $own_http == 200 ]]||fail "applicant cancellation expected HTTP 200 got $own_http"
[[ $own_state == 'cancelled|1|1' ]]||fail "own approval expected cancelled with reviewer cleared and completion time got $own_state"
[[ $history_actor == "$admin" && $audit_actor == "$admin" ]]||fail "own cancellation history/audit actor expected admin got $history_actor/$audit_actor"
[[ $foreign_http == 400 ]]||fail "non-applicant cancellation expected HTTP 400 got $foreign_http"
[[ $foreign_state == "pending|$other" ]]||fail "foreign approval must remain pending for original applicant got $foreign_state"
[[ $foreign_history == 0 && $foreign_audit == 0 ]]||fail "rejected foreign cancellation must not create history/audit got $foreign_history/$foreign_audit"
finish 6
