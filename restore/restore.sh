#!/bin/bash
set -e

SOURCE_PROJECT_NAME="project-rekor"
SOURCE_SQL_INSTANCE_NAME="rekor-dev"

TARGET_PROJECT_NAME="openssf"
TARGET_SQL_INSTANCE_NAME="rekor-5-7"

BACKUP_LIST=$(mktemp)

# Fetches the list of backups on the source instance
echo "Fetching list of backups available from the source SQL instance '$SOURCE_SQL_INSTANCE_NAME' in project '$SOURCE_PROJECT_NAME' ..."
curl -f -s -o $BACKUP_LIST -X GET -H "Authorization: Bearer $(gcloud auth print-access-token)" "https://sqladmin.googleapis.com/v1/projects/${SOURCE_PROJECT_NAME}/instances/${SOURCE_SQL_INSTANCE_NAME}/backupRuns"

# Calculates the backup run ID for the most recent successful backup
BACKUP_RUN_ID=$(jq -r '[.items[] | select(.status | contains("SUCCESSFUL"))] | sort_by(.startTime) | reverse | .[0].id' $BACKUP_LIST)

# Prints the time the backup was taken
START_TIME_BACKUP_RUN_ID=$(jq -r '[.items[] | select(.id | contains("'$BACKUP_RUN_ID'"))] | .[0].startTime ' $BACKUP_LIST)
echo "Attempting to restore backup run $BACKUP_RUN_ID from $START_TIME_BACKUP_RUN_ID into target SQL instance '$TARGET_SQL_INSTANCE_NAME' in project '$TARGET_PROJECT_NAME' ..."

generate_restore_request()
{
  cat <<EOF
{
  "restoreBackupContext":
  {
    "backupRunId": "$BACKUP_RUN_ID",
    "project": "${SOURCE_PROJECT_NAME}",
    "instanceId": "${SOURCE_SQL_INSTANCE_NAME}"
  }
}
EOF
}

BACKUP_OPERATION_INFO=$(mktemp)
# Sends the restore from backup request to the target instance within the target project
curl -f -s -o $BACKUP_OPERATION_INFO -X POST -H "Authorization: Bearer $(gcloud auth print-access-token)" -H "Content-Type: application/json; charset=utf-8" --data "$(generate_restore_request)" "https://sqladmin.googleapis.com/v1/projects/${TARGET_PROJECT_NAME}/instances/${TARGET_SQL_INSTANCE_NAME}/restoreBackup"

# Block returning from this script until operation completes
OPERATION_NAME=$(jq -r '.name' $BACKUP_OPERATION_INFO)
gcloud sql operations wait "$OPERATION_NAME" --timeout=unlimited --project=$TARGET_PROJECT_NAME
