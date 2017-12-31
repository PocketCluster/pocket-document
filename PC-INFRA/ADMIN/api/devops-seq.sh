# ---------------- FQDN ----------------
# api.pocketcluster.io

# ---------------- show installed packages ----------------
apt list --installed

# ---------------- /etc/hostname ---------------- 
echo "api" > /etc/hostname

# ---------------- /etc/hosts ---------------- 
# Your system has configured 'manage_etc_hosts' as True.
# As a result, if you wish for changes to this file to persist
# then you will need to either
# a.) make changes to the master file in /etc/cloud/templates/hosts.tmpl
# b.) change or remove the value of 'manage_etc_hosts' in
#     /etc/cloud/cloud.cfg or cloud-config from user-data
#
127.0.0.1 localhost

# The following lines are desirable for IPv6 capable hosts
::1 ip6-localhost ip6-loopback
fe00::0 ip6-localnet
ff00::0 ip6-mcastprefix
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
ff02::3 ip6-allhosts

127.0.1.1       api.pocketcluster.io api
107.170.237.110 api.pocketcluster.io api

# ---------------- Create User ---------------- 
addgroup --gid 29999 pocketman &&\
adduser --gecos "PocketCluster (admin user)" --add_extra_groups --disabled-password --gid 29999 --uid 29999 poimdall &&\
usermod -a -G sudo -p $(mkpasswd -m sha-512 pocket $(date +%m%H%M%S)) poimdall &&\
echo "poimdall ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/poimdall

mkdir -p /home/poimdall/.ssh
echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDxxtLwF2sp8lONgTZ9uDy3UwBGADpJl0HxjclNrPO3wqXvh2qlvSzVqf7WK8BVaVlMadMfGgYXCkjJy7O+Eje2RhdUjDni2BLZMGBaAefQom6H9QNk5hX16xJJU2080KZORl4dB6/aMEQ32kDgEAZWDiG0nbJlUGjvLpzJFDqWK98za70gNKtpr1sGELGM6Fvx4J7EF05zGTFFfX0ZP+w6K0CD3i5AHfHxwHdpkJ9jp1XUwb0cBogsxbtdIoX7RzyeEDLBoNTR8fUg5AFuwLymCJ7ozXMgeviuXSInUR5Jc6YYOBlO0MetJ648f7icYIvi9eUIUL6Ds60+FToLwjn9 stkim1@colorfulglue.com" > /home/poimdall/.ssh/authorized_keys
echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDbzPpw3fJ89KiVpJxO73xlFXZZkvi6qrs+Tm/FUPQ0YPvKLOIN9W6mU5y4zoAZPXuJD1WHDFSWKbgmsvwAbS+MBH0Cf/3JBq8peYv66pWiuG+eoKJdm+SfPph24GL/Va9H4iWl4MvabCCX7t7GmrNoO/CGSFtDILdlKrmnVa3dc/9+KKKwD2AtlAG0hFp/ECK9uEb1AZhTFUaQ9r4yIAdLWwB9keTaylnoKDsH1N/INjRC4KnC0/im/Pg02y9oW2yFIRjUso3Kre43flnoemPnfbzIE054O6B/10YxDjCctXIWUGa41D3zL1tER2G9+v0xZJjrdufllbq9lqOQmNhZ stkim1@vpn-sf" >> /home/poimdall/.ssh/authorized_keys
chown -R poimdall:pocketman /home/poimdall/.ssh && chmod 0700 /home/poimdall/.ssh/ && chmod 0600 /home/poimdall/.ssh/*

# ---------------- Install Packages ---------------- 
apt update && apt upgrade -y && apt install --no-install-suggests -y dialog language-pack-en-base libpam-systemd dbus fail2ban logwatch iptables-persistent apparmor apparmor-profiles git

# ---------------- Set Locale ---------------- 
cat <<EOM >/etc/default/locale
LANG="en_US.UTF-8"
LANGUAGE="en_US.UTF-8"
LC_NUMERIC="en_US.UTF-8"
LC_TIME="en_US.UTF-8"
LC_MONETARY="en_US.UTF-8"
LC_PAPER="en_US.UTF-8"
LC_NAME="en_US.UTF-8"
LC_ADDRESS="en_US.UTF-8"
LC_TELEPHONE="en_US.UTF-8"
LC_MEASUREMENT="en_US.UTF-8"
LC_IDENTIFICATION="en_US.UTF-8"
LC_CTYPE="UTF-8"
LC_COLLATE="en_US.UTF-8"
LC_ALL="en_US.UTF-8"
EOM

cat <<EOM >/etc/locale.gen
# This file lists locales that you wish to have built. You can find a list
# of valid supported locales at /usr/share/i18n/SUPPORTED, and you can add
# user defined locales to /usr/local/share/i18n/SUPPORTED. If you change
# this file, you need to rerun locale-gen.

en_US.UTF-8 UTF-8
EOM

locale-gen en_US.UTF-8
update-locale LC_ALL=en_US.UTF-8

# ---------------- Set Timezone ---------------- 
# https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
echo "America/Los_Angeles" > /etc/timezone 
dpkg-reconfigure tzdata

# ---------------- Setup SSHD ---------------- 
vi /etc/ssh/sshd_config

Protocol 2
PermitRootLogin no
PermitEmptyPasswords no
PasswordAuthentication no
AuthorizedKeysFile      %h/.ssh/authorized_keys
AllowUsers poimdall

# ---------------- Setup Fail2Ban ---------------- 
bantime = 86400
destemail = root@localhost, stkim1@colorfulglue.com
sender = root@api.pocketcluster.org
port = 0:65535
banaction = iptables-multiport

scp jail.local api:~/
scp filter.d/* api:~/

(mv jail.local /etc/fail2ban/ || true) && (mv *.conf /etc/fail2ban/filter.d || true) && chown -R root:root /etc/fail2ban/*
touch /var/log/sshd.log 

service fail2ban restart

# force postfix to use ipv4 only
# http://www.postfix.org/IPV6_README.html
vi /etc/postfix/main.cf
inet_protocols = ipv4
service postfix restart

# ---------------- Setup Logwatch ---------------- 
scp 00logwatch api:~/
mv 00logwatch /etc/cron.daily/ && chown root:root /etc/cron.daily/*

# ---------------- Setup IpTables ---------------- 
scp rules.v4 api:~/
mv rules.v4 /etc/iptables/ && chown root:root /etc/iptables/*
iptables-apply /etc/iptables/rules.v4

# ---------------- Kernel Param Tuning ---------------- 
# show all parameters
$ sysctl -a

# Source routing is an Internet Protocol mechanism that allows an IP packet to 
# carry information, a list of addresses, that tells a router the path the packet
# must take. There is also an option to record the hops as the route is traversed.
# The list of hops taken, the "route record", provides the destination with a 
# return path to the source. This allows the source (the sending host) to specify 
# the route, loosely or strictly, ignoring the routing tables of some or all of 
# the routers. It can allow a user to redirect network traffic for malicious purposes. 
# Therefore, source-based routing should be disabled.

cat <<EOM >/etc/sysctl.conf
# Disble IPv6
net.ipv6.conf.all.disable_ipv6 = 1
net.ipv6.conf.default.disable_ipv6 = 1
net.ipv6.conf.lo.disable_ipv6 = 1

# Increase number of incoming connections that can queue up before dropping
# The maximum number of "backlogged sockets".  Default is 128.
net.core.somaxconn = 2048

# Disable packet forwarding.
net.ipv4.ip_forward = 0
net.ipv6.conf.all.forwarding = 0

# IP Spoofing protection
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1

# Ignore ICMP broadcast requests
net.ipv4.icmp_echo_ignore_broadcasts = 1

# Disable source packet routing
net.ipv4.conf.all.accept_source_route = 0
net.ipv6.conf.all.accept_source_route = 0 
net.ipv4.conf.default.accept_source_route = 0
net.ipv6.conf.default.accept_source_route = 0

# Ignore send redirects (we are not a router)
net.ipv4.conf.all.send_redirects = 0
net.ipv4.conf.default.send_redirects = 0

# Block SYN attacks
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_max_syn_backlog = 2048
net.ipv4.tcp_synack_retries = 2
net.ipv4.tcp_syn_retries = 5

# Log Martians
net.ipv4.conf.all.log_martians = 1
net.ipv4.icmp_ignore_bogus_error_responses = 1

# Ignore ICMP redirects (prevent MITM attacks)
net.ipv4.conf.all.accept_redirects = 0
net.ipv6.conf.all.accept_redirects = 0
net.ipv4.conf.default.accept_redirects = 0 
net.ipv6.conf.default.accept_redirects = 0

# Ignore Directed pings
net.ipv4.icmp_echo_ignore_all = 1
EOM

# reload!
sysctl -p

# ----------------- NGINX(#1) Let's Encrypt ---------------------
 git clone https://github.com/certbot/certbot /opt/letsencrypt

# Stop Cloudflare for the first time
service nginx stop

cd /opt/letsencrypt
./letsencrypt-auto certonly --standalone --email stkim1@colorfulglue.com -d api.pocketcluster.io

: <<OUTPUT
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Obtaining a new certificate
Performing the following challenges:
tls-sni-01 challenge for api.pocketcluster.io
Waiting for verification...
Cleaning up challenges

IMPORTANT NOTES:
 - Congratulations! Your certificate and chain have been saved at:
   /etc/letsencrypt/live/api.pocketcluster.io/fullchain.pem
   Your key file has been saved at:
   /etc/letsencrypt/live/api.pocketcluster.io/privkey.pem
   Your cert will expire on 2017-11-19. To obtain a new or tweaked
   version of this certificate in the future, simply run
   letsencrypt-auto again. To non-interactively renew *all* of your
   certificates, run "letsencrypt-auto renew"
 - If you like Certbot, please consider supporting our work by:

   Donating to ISRG / Let's Encrypt:   https://letsencrypt.org/donate
   Donating to EFF:                    https://eff.org/donate-le
OUTPUT

# ----------------- NGINX(#1) Config ---------------------
mkdir -p /etc/nginx/ssl.crt/
openssl dhparam -out /etc/nginx/ssl.crt/dhparams.pem 2048;

mkdir -p /api-server && chown www-data:www-data /api-server && chmod 775 /api-server
usermod -a -G www-data poimdall

scp default api:~/
scp api.pocketcluster.io.conf api:~/
scp nginx.conf api:~/

mv default /etc/nginx/sites-available/
mv api.pocketcluster.io.conf /etc/nginx/conf.d/
mv nginx.conf /etc/nginx/

chown root:root /etc/nginx/sites-available/*
chown root:root /etc/nginx/conf.d/*
chown root:root /etc/nginx/nginx.conf

nginx -c /etc/nginx/nginx.conf -t
service nginx restart && service nginx status

# ------------ Secure shared memory. -------------------
# error : [    9.969353] cgroup: new mount options do not match the existing superblock, will be ignored ???
#echo "tmpfs     /run/shm  tmpfs  defaults,noexec,nosuid  0 0" >> /etc/fstab


# ------------ API Service ------------
mkdir -p /api-service && chown www-data:www-data /api-service && chmod 775 /api-service
touch /var/log/api-service.log && chown root:adm /var/log/api-service.log && chmod 640 /var/log/api-service.log
usermod -a -G www-data poimdall
usermod -aG adm poimdall

scp api-service api:~/
scp api-service.service api:~/

mv api-service /usr/local/bin/ && chown root:root /usr/local/bin/api-service
mv api-service.service /etc/systemd/system/ && chown root:root /etc/systemd/system/api-service.service

systemctl daemon-reload
systemctl start api-service
systemctl enable api-service
systemctl status api-service.service
