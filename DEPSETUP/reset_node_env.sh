#!/usr/bin/env bash

# This script assume that
# 1) all the hosts can be accessed with sshkey/recvtest.pem
# 2) all the hosts have `pocket` account with admin priviledge
# 3) all the hosts have the same version of go and "apt install build-essential" installed

set -ex
SSHKEY="./sshkey/recvtest.pem"
TARGET_HOST=${1}

if [[ -z ${TARGET_HOST} ]]; then
    echo "unable to coninue w/o target host"
fi

function reset_ssh_command() {
    HOST=${1}
    CMD=${2}

    # ssh -i ${SSHKEY} ${HOST} "${CMD}"
    ssh ${HOST} "${CMD}"
}

function start_pocket() {
    echo "sudo -s service pocket start"
}

function stop_pocket() {
    echo "sudo -s service pocket stop"
}

function cmd_reset_hostname() {
    HOST=${1}
    echo "echo ${HOST} | sudo -s tee /etc/hostname"
}

function cmd_remove_pocket() {
    echo "sudo -s rm -rf /etc/pocket"
}

function recover_cert_auth() {
    HOST=${1}
    scp ./sshkey/ca-certificates.crt ${HOST}:~/
    echo "sudo -s chown root:root ca-certificates.crt && sudo -s chmod 644 ca-certificates.crt && sudo -s mv ca-certificates.crt /etc/ssl/certs/"
}

reset_ssh_command ${TARGET_HOST} "$(stop_pocket)"
reset_ssh_command ${TARGET_HOST} "$(cmd_reset_hostname ${TARGET_HOST})"
reset_ssh_command ${TARGET_HOST} "$(cmd_remove_pocket)"
reset_ssh_command ${TARGET_HOST} "$(recover_cert_auth ${TARGET_HOST})"
reset_ssh_command ${TARGET_HOST} "$(start_pocket)"
