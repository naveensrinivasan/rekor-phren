#!/bin/bash
set -e

SOURCE_PROJECT_NAME="project-rekor"
SOURCE_SQL_INSTANCE_NAME="rekor-dev"

TARGET_PROJECT_NAME="openssf"
TARGET_SQL_INSTANCE_NAME="rekor-5-7"

# Need to get the list of recent backups
BACKUP_LIST=$(mktemp)

# Fetches the list of backups on the source instance
echo "Fetching list of backups available from the source instance..."
curl -f -s -o $BACKUP_LIST -X GET -H "Authorization: Bearer $(gcloud auth print-access-token)" "https://sqladmin.googleapis.com/v1/projects/${SOURCE_PROJECT_NAME}/instances/${SOURCE_SQL_INSTANCE_NAME}/backupRuns"

# Calculates the backup run ID for the most recent successful backup
BACKUP_RUN_ID=$(jq -r '[.items[] | select(.status | contains("SUCCESSFUL"))] | sort_by(.startTime) | reverse | .[0].id' $BACKUP_LIST)

# Prints the time the backup was taken
START_TIME_BACKUP_RUN_ID=$(jq -r '[.items[] | select(.id | contains("'$BACKUP_RUN_ID'"))] | .[0].startTime ' $BACKUP_LIST)
echo "Attempting to restore backup run $BACKUP_RUN_ID from $START_TIME_BACKUP_RUN_ID..."

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

# Sends the restore from backup request to the target instance within the target project
curl -f -X POST -H "Authorization: Bearer $(gcloud auth print-access-token)" -H "Content-Type: application/json; charset=utf-8" --data "$(generate_restore_request)" "https://sqladmin.googleapis.com/v1/projects/${TARGET_PROJECT_NAME}/instances/${TARGET_SQL_INSTANCE_NAME}/restoreBackup"
