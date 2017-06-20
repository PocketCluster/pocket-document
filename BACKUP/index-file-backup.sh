#/usr/bin/env bash

DATE=$(date +%F)
IDFILE_PATH="/home/poimdall/.ssh/id_rsa_mbp.pub"
SOURCE_PATH="index:/www-server"
BACKUP_PATH="/work/backup/index/${DATE}"

if [[ ! -d ${BACKUP_PATH} ]]; then
    mkdir -p ${BACKUP_PATH}
fi
if [[ ! -d "${BACKUP_PATH}/readme/" ]]; then
    mkdir -p "${BACKUP_PATH}/readme/"
fi

#scp -i ${IDFILE_PATH} "${SOURCE_PATH}/pc-index.db" "${BACKUP_PATH}"
#scp -i ${IDFILE_PATH} "${SOURCE_PATH}/*.bt"        "${BACKUP_PATH}"

rsync -rlmvt -e "ssh -o ProxyCommand='ssh -i /home/poimdall/.ssh/id_rsa_mbp.pub stkim1@vpn-sf.pocketcluster.io nc %h %p' stkim1@index" "${SOURCE_PATH}/readme/*" "${BACKUP_PATH}/readme/"