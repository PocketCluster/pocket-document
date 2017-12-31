#!/usr/bin/env bash

export SENDY_UPGRADE=$1
if [[ -z ${SENDY_UPGRADE} ]]; then
	echo "[ERROR] upgrade file must be specified"
	exit
fi
if [[ ! -f ${SENDY_UPGRADE} ]]; then
    echo "[ERROR] the specified upgrade file must exists"
    exit
fi
if [ $(id -u) -ne 0 ]; then
    echo "[ERROR] Must be root."
    exit 1
fi

echo "Extract ${SENDY_UPGRADE}"
unzip ${SENDY_UPGRADE}
echo "Copy old config..."
cp /www/sendy/mailer/includes/config.php ${PWD}/sendy/includes/config.php
cat ${PWD}/sendy/includes/config.php
sleep 1

echo "Copy old uploads..."
cp -rf /www/sendy/mailer/uploads/* 		 ${PWD}/sendy/uploads/
ls ${PWD}/sendy/uploads/
sleep 1

echo "Change ownership and name..."
mv ${PWD}/sendy ${PWD}/mailer
chown -R www-data:www-data ${PWD}/mailer/

if [[ -d /www/sendy/mailer ]]; then
	echo "old directory exists. backing it up"
	mv /www/sendy/mailer /www/sendy/mailer-old
fi
mv ${PWD}/mailer /www/sendy/mailer