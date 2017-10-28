#!/usr/bin/env bash

BACKUP_DATE=$(date +"%Y-%m-%d")
BACKUP_FOLDER=${BACKUP_FOLDER:-"/work/backup/news/${BACKUP_DATE}"}

if [ -f "${BACKUP_FOLDER}/sendy-backup.sql" ]; then
    exit
fi

/bin/mkdir -p "${BACKUP_FOLDER}" && \
/usr/bin/ssh news "mysqldump -u sendy -psendy sendy > /home/stkim1/sendy-backup-${BACKUP_DATE}.sql" && \
/usr/bin/scp news:/home/stkim1/sendy-backup-${BACKUP_DATE}.sql "${BACKUP_FOLDER}/sendy-backup.sql" && \
/usr/bin/ssh news "rm /home/stkim1/sendy-backup-${BACKUP_DATE}.sql"