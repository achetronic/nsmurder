#!/usr/bin/env bash
set -eo pipefail

# This script is intended to solve all the little issues not solved before deleting EKS
# Add here all needed functions to assure that EKS deletion is done successfully after this execution
# WARNING: All the pipeline depends on this script so be extremely careful with the things you do


# This path is defined to avoid relative-pathing issues
SELF_PATH=$(cd $(dirname "$0"); pwd)

# Create graphical timers
function timer () {

  local SECS=$1;
  local MSG=$2

  printf "[···] %s ... %02d:%02d\n" "${MSG}" $(( (SECS/60)%60)) $((SECS%60))
  sleep "${SECS}"
  printf "[ok] %s \n" "${MSG}"
}

# Schedule deletion for all namespaces
# WARNING: There are several ignored namespaces because we need them until all the process finishes
function schedule_namespaces_deletion () {
  echo -e "[···] Scheduling namespaces deletion"

  local IGNORE_NAMESPACES="kube-system,kube-public,kube-node-lease,default,external-dns,calico-system"
  local IGNORED_NAMESPACES_STRING=""

  # Craft a list ready for kubectl
  for namespace in $(echo $IGNORE_NAMESPACES | tr ',' '\n')
  do
    IGNORED_NAMESPACES_STRING="$IGNORED_NAMESPACES_STRING,metadata.name!=$namespace"
  done

  # Filter unwanted chars from the list
  IGNORED_NAMESPACES_STRING=$(echo "$IGNORED_NAMESPACES_STRING" | sed --regexp-extended 's/^[\,]//')

  kubectl delete namespace \
    --field-selector "$IGNORED_NAMESPACES_STRING" \
    --force=true \
    --grace-period=0 \
    --wait=false 2>/dev/null || (echo "[X] Failed deletion" && exit 1)

  echo "[ok] Deletion schedule success"

  timer 120 "Waiting prudential time before using brute force"
}

# Delete unavailable apiservices
# REF: More info in https://github.com/kubernetes/kubernetes/issues/60807#issuecomment-524772920
function delete_unavailable_apis () {
  echo -e "[···] Deleting unavailable APIs"

  local APIS=$(kubectl get apiservice | grep False | cut -f1 -d ' ')

  for API in $APIS; do
    echo "[···] Deleting broken API: ${API}"

    timeout 60 kubectl delete apiservice "${API}" >& /dev/null || echo "[X] Failed deletion"

    echo "[ok] Deletion success"
  done
}

# Delete all resources inside stuck namespaces
function delete_stuck_namespace_resources () {
  echo -e "[···] Deleting resources inside stuck namespaces"
  echo -e "[INFO] Remember some resources depends on others and are deleted automatically after deleting parents"

  local API_RESOURCES=$(kubectl api-resources --verbs=list --namespaced --output=name 2>/dev/null)
  local STUCK_NAMESPACES=$(kubectl get ns --field-selector status.phase=Terminating -o jsonpath='{.items[*].metadata.name}')
  local NAMESPACE_RESOURCES

  # Add 'default' namespace to the list of namespaces to clean
  STUCK_NAMESPACES="${STUCK_NAMESPACES} default"

  for ns in $STUCK_NAMESPACES; do

    echo "[···] Getting the resources in namespace: ${ns}"
    # Get all the resources on that namespace and filter them
    NAMESPACE_RESOURCES=$(
      echo "${API_RESOURCES}" | \
      xargs -n 1 \
        kubectl get \
          --namespace="${ns}" \
          --no-headers=true \
          --output=jsonpath="{range .items[*]}{.kind}/{.metadata.name}{'\n'}{end}" 2>/dev/null | \
      sed -E "s#List/##gI"
    )

    echo "[···] Delete the stuck resources in namespace: ${ns}"
    while read -r resource; do
      # Mark all resources for deletion inside the namespace
      # WARNING: This step is done because of 'default' namespace can not be directly deleted,
      # so NOT marked as Terminating. Schedule resource deletion always work
      kubectl delete "${resource}" \
          --namespace="${ns}" \
          --grace-period=0 \
          --force=true \
          --wait=false 2>/dev/null \
          || (echo "[X] Some resource could not be deleted: ${resource}" && exit 0)

      # Patch all resources finalizers
      kubectl patch "${resource}" \
          --namespace="${ns}" \
          --type=merge \
          --patch='{"metadata":{"finalizers":null}}' 2>/dev/null \
          || (echo "[X] Some resource could not be patched: ${resource}" && exit 0)
    done <<< "${NAMESPACE_RESOURCES}"
  done

  echo "[ok] Deletion success"
  timer 120 "Waiting prudential time before using brute force"
}

# Force the namespace deletion to delete the resources inside
# This is the last resource to clean the cluster before destroying it
function delete_stuck_namespaces () {
  echo -e "[···] Deleting stuck namespaces (with force)"

  local EXIT_CODES=()

  for ns in $(kubectl get ns --field-selector status.phase=Terminating -o jsonpath='{.items[*].metadata.name}'); do

    kubectl get ns "${ns}" -ojson | \
      jq '.spec.finalizers = []' | \
      kubectl replace --raw "/api/v1/namespaces/$ns/finalize" -f -;

    # Add the exit code of the previous command
    EXIT_CODES+=($?)
  done

  # Review all exit codes
  # WARNING: All commands must exit with a code = 0 to consider patching as a success
  for exit_code in "${EXIT_CODES[@]}"; do
    if [[ $exit_code -gt 0 ]]; then
      echo "[X] Failed forced deletion"
      exit 1;
      break
    fi
  done

  echo "[ok] Deletion success"
}


# Run the script
{
  echo -e "\n\n======== Cluster cleaning ========\n\n"

  # Script checks
  if ! command -v flux &> /dev/null; then
      echo "[X] This script need Flux installed to work."
      exit 1
  fi

  if ! command -v kubectl &> /dev/null; then
      echo "[X] This script need Kubectl installed to work."
      exit 1
  fi

  echo "[...] Executing script using CLI versions: "
  echo "FluxCD: $(flux --version)"
  echo "Kubectl: $(kubectl version --output=json )"

  echo "-------------------"
  echo "[ok] All checks passed. Starting the script"

  schedule_namespaces_deletion
  delete_unavailable_apis
  delete_stuck_namespace_resources
  delete_stuck_namespaces
}
