#!/usr/bin/env bash

#set -ex

export NODE_NAME=$1
export IP_ADDR=$2

function generate_cert_key_pair() {
	if [[ -z ${NODE_NAME} ]]; then
		echo "!!!Empty nodename!!!"
	fi
	if [[ -z ${IP_ADDR} ]]; then
		echo "!!!invalid ip address!!!"
	fi
	local SED=$(type -p sed)
	local OPENSSL=$(type -p openssl)
	echo "${NODE_NAME} / ${IP_ADDR}"

	SED s/HOST_NAME/${NODE_NAME}/ conf.template > ${NODE_NAME}.conf
	SED -i '' s/IP_ADDRESS/${IP_ADDR}/ ${NODE_NAME}.conf

	OPENSSL genrsa -out ${NODE_NAME}.key 2048 -sha256
	OPENSSL req -new -key ${NODE_NAME}.key -out ${NODE_NAME}.csr -sha256 -subj "/CN=${NODE_NAME}" -config ${NODE_NAME}.conf
	OPENSSL x509 -req -in ${NODE_NAME}.csr -CA ca-cert.pub -CAkey ca-key.pem -CAcreateserial -out ${NODE_NAME}.cert -sha256 -days 365 -extensions v3_req -extfile ${NODE_NAME}.conf
}

generate_cert_key_pair