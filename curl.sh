# state variables
instance_id="d226248e-a599-4b33-a9c2-bd0790960abc"
service_id="06948cb0-cad7-4buh-leba-9ed8b5c345a3"
plan_id="7c0b2254-7e68-11e7-bbe1-000d3a818256"
binding_id="532cc413-2a8b-4e59-822c-14bd4e13e2aa"

# define functions
generate_provision_post_data(){
cat <<EOF
{
"context": {
  "platform": "cloudfoundry"
},
"service_id": "${service_id}",
"plan_id": "${plan_id}",
"organization_guid": "4a0ba950-3509-4098-8445-6c1bbc5f9229",
"space_guid": "af746f9a-b944-4830-a2ae-25ccea2635d7"
}
EOF
}

generate_bind_post_data(){
cat <<EOF
{
  "service_id": "${service_id}",
  "plan_id": "${plan_id}"
}
EOF
}

echo "start to test service broker\n"
# echo "==============================catalog=============================="
# curl "http://admin:admin@127.0.0.1:9000/v2/catalog"
# echo "==============================provision=============================="
# curl -H "Content-Type:application/json" \
# -X PUT \
# -d "$(generate_provision_post_data)" "http://admin:admin@127.0.0.1:9000/v2/service_instances/$instance_id?accepts_incomplete=true"
# echo "==============================last_operation=============================="
# curl "http://admin:admin@127.0.0.1:9000/v2/service_instances/$instance_id/last_operation?operation=provision:$instance_id"
# echo "sleep 5 min for provision"
# sleep 5m 
# echo "check status second time"
# curl "http://admin:admin@127.0.0.1:9000/v2/service_instances/$instance_id/last_operation"
echo "==============================bind=============================="
curl -H "Content-Type:application/json" \
-X PUT \
-d "$(generate_bind_post_data)" "http://admin:admin@127.0.0.1:9000/v2/service_instances/$instance_id/service_bindings/$binding_id"
echo "==============================unbind=============================="
curl -X DELETE "http://admin:admin@127.0.0.1:9000/v2/service_instances/$instance_id/service_bindings/$binding_id?service_id=$service_id&plan_id=plan_id"
echo "==============================deprovisioning=============================="
curl -X DELETE "http://admin:admin@127.0.0.1:9000/v2/service_instances/$instance_id?accepts_incomplete=true&service_id=$service_id&plan_id=plan_id"
echo "==============================last_operation=============================="
curl "http://admin:admin@127.0.0.1:9000/v2/service_instances/$instance_id/last_operation?operation=deprovision:$instance_id"
echo "sleep 2 min for deprovision"
sleep 2m 
echo "check status second time"
curl "http://admin:admin@127.0.0.1:9000/v2/service_instances/$instance_id/last_operation?operation=provision:$instance_id"
