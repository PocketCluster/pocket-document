#!/usr/bin/env bash

# This script assume that
# 1) all the hosts can be accessed with sshkey/recvtest.pem
# 2) all the hosts have `pocket` account with admin priviledge
# 3) all the hosts have the same version of go and "apt install build-essential" installed

set -ex
SSHKEY="./sshkey/recvtest.pem"

function reset_ssh_command() {
    HOST=${1}
    CMD=${2}

    ssh -i ${SSHKEY} ${HOST} "${CMD}"
}

function cmd_reset_hostname() {
    HOST=${1}
    echo "sudo -s rm -rf /etc/pocket/ && echo ${HOST} | sudo -s tee /etc/hostname"
}

reset_ssh_command pocket@192.168.1.161 "$(cmd_reset_hostname pc-node1)"
reset_ssh_command pocket@192.168.1.162 "$(cmd_reset_hostname pc-node2)"
reset_ssh_command pocket@192.168.1.163 "$(cmd_reset_hostname pc-node3)"
reset_ssh_command pocket@192.168.1.164 "$(cmd_reset_hostname pc-node4)"
reset_ssh_command pocket@192.168.1.165 "$(cmd_reset_hostname pc-node5)"
reset_ssh_command pocket@192.168.1.166 "$(cmd_reset_hostname pc-node6)"

rm -rf ${HOME}/.pocket/*