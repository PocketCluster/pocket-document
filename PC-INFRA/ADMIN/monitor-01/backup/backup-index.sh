#!/usr/bin/env bash

BACKUP_DATE=$(date +"%Y-%m-%d")
BACKUP_FOLDER=${BACKUP_FOLDER:-"/work/backup/index/${BACKUP_DATE}"}

if [ -d "${BACKUP_FOLDER}" ] || [ -f "/work/backup/index/pc-archive-${BACKUP_DATE}.tar.gz" ]; then
    exit
fi

/bin/mkdir -p "${BACKUP_FOLDER}" && \
/usr/bin/rsync -rlmvtz -e ssh --exclude=indexsrv --exclude=htmlscrap index:/www-server/* "${BACKUP_FOLDER}" && \
cd "${BACKUP_FOLDER}" && \
/bin/tar -cvzf "/work/backup/index/pc-archive-${BACKUP_DATE}.tar.gz" ./* && \
cd .. && rm -rf "${BACKUP_FOLDER}"
