# ---------------- FQDN ----------------
# us-west-01.cdn.pocketcluster.io

# ---------------- /etc/hostname ---------------- 
#echo "us-west-01" > /etc/hostname

# ---------------- /etc/hosts ---------------- 
#ff02::1        ip6-allnodes
#ff02::2        ip6-allrouters
#::1            localhost ip6-localhost ip6-loopback

127.0.1.1 us-west-01.cdn.pocketcluster.io us-west-01
127.0.0.1 localhost.localdomain localhost
# Auto-generated hostname. Please do not remove this comment.
107.181.138.177 us-west-01.cdn.pocketcluster.io  us-west-01
::1 localhost

# ---------------- Create User ---------------- 
addgroup --gid 29999 pocketman &&\
adduser --gecos "PocketCluster (admin user)" --add_extra_groups --disabled-password --gid 29999 --uid 29999 poimdall &&\
usermod -a -G sudo -p $(mkpasswd -m sha-512 NONONPASSWD $(date +%m%H%M%S)) poimdall &&\
echo "poimdall ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/poimdall

# ssh-keygen -t rsa -b 4096 -C "stkim1@pocketcluster.io"
echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDRObM1vdIRjM0q7MoexT7ZY2tI/3wfzMwOcFf9A8RbIv3x6meUcmXkOUe33QheRT+w7vHevZBBO4yke4h4jFC3F1LL6UOIBAiSEZaJc/bVq2hK0UmD7EzMD0+VrrvRi/QgMS0e6lH+qp2tQRU6rBNjY0n1DuTYXh0JCck2phkuRwGpy8O7L9BLJfKZ/F2fKXLVdIjUan2xS+mmnp5QvPnXA5fWbCDFnBZY7Sor/Iq4CmEY41tpYtklOn4vfP8w3E/f+Yb0V9TcIkJzJxUsgtKUKU7cLvx8Tl0lUda2ql/JfP0hjBt48TnDAAF3pYF7FNb8V81FGit4/g60kzf+2MB662rru3Y2Md21qOwGsLmSSQeM2UuCycE+SWVzl47XpaZdem+MbNA7crhe85b7opyKylW2pKnWrlz6YUdpR788nMpON6oOBmHNYqVI1scvlP9qoighMGKoNT9fQb3qYEEmfhqd9RYrXi542XJ+JypX37rmZjOmM29ay2Z9VcRbgNrzyh2R6BywHABxPY/Uo5eNHBtAqq0L1YDAB417SFH6RHgD5invyVD5i5uS3S2OmkpIjyb6T3qgJiDG13H5K1+D8vJbLP14vnIeWuvsbADnzE6O+zBf/cv8L4iZhtN172wGcmRzjJU4h4hXWmUnD4WKcVbOzx0tqFZaH7qzUgFdNQ== stkim1@pocketcluster.io" > ~poimdall/.ssh/authorized_keys
chown -R poimdall:pocketman ~poimdall/.ssh && chmod 0700 ~poimdall/.ssh/ && chmod 0600 ~poimdall/.ssh/*

# ---------------- Install Packages ---------------- 
apt update && apt upgrade -y && apt install --no-install-suggests -y dialog language-pack-en-base libpam-systemd dbus fail2ban logwatch iptables-persistent git apparmor apparmor-profiles

# ---------------- Set Timezone ---------------- 
# https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
echo "America/Los_Angeles" > /etc/timezone 
dpkg-reconfigure -f noninteractive tzdata

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

# ---------------- Setup SSHD ---------------- 
vi /etc/ssh/sshd_config

PermitRootLogin no
PermitEmptyPasswords no
PasswordAuthentication no
AuthorizedKeysFile      %h/.ssh/authorized_keys
AllowUsers poimdall

# ---------------- Setup Fail2Ban ---------------- 
bantime = 86400
destemail = root@localhost, stkim1@colorfulglue.com
sender = root@us-west-01.cdn.pocketcluster.io
port = 0:65535
banaction = iptables-multiport

scp jail.local us-west-01:~/
scp filter.d/* us-west-01:~/

(mv jail.local /etc/fail2ban/ || true) && (mv *.conf /etc/fail2ban/filter.d || true) && chown -R root:root /etc/fail2ban/*
service fail2ban restart

# force postfix to use ipv4 only
# http://www.postfix.org/IPV6_README.html
vi /etc/postfix/main.cf
inet_protocols = ipv4
service postfix restart

# ---------------- Setup Logwatch ---------------- 
scp 00logwatch us-west-01.cdn.pocketcluster.io:~/
mv 00logwatch /etc/cron.daily/ && chown root:root /etc/cron.daily/*

# ---------------- Setup IpTables ---------------- 
mkdir -p /etc/iptables
scp rules.v4 us-west-01.cdn.pocketcluster.io:~/
mv rules.v4 /etc/iptables/ && chown root:root /etc/iptables/*
iptables-apply /etc/iptables/rules.v4

# ---------------- Reboot ---------------- 
reboot

cat /proc/sys/net/ipv6/conf/all/disable_ipv6
1

# ---------------- Remove Packages ---------------- 
apt-get remove apache2* samba-common
apt list --installed


```????
auto venet0:0
iface venet0:0 inet static
    address 107.181.138.177
    netmask 255.255.255.255
```

# ---------------- CDN Configuration ---------------- 
mkdir -p /cdn-content && chown www-data:www-data /cdn-content && chmod 775 /cdn-content
usermod -a -G www-data poimdall

scp cdn_pocketcluster_org.crt us-west-01:~/
scp cdn_pocketcluster_org.key us-west-01:~/
scp default us-west-01:~/
scp cdn.pocketcluster.io.conf us-west-01:~/
scp nginx.conf us-west-01:~/

mkdir -p /etc/nginx/ssl 
openssl dhparam -out /etc/nginx/ssl/dhparams.pem 2048
mv default /etc/nginx/sites-available/
mv cdn_pocketcluster_org.key /etc/nginx/ssl/
mv cdn_pocketcluster_org.crt /etc/nginx/ssl/
chown root:root /etc/nginx/ssl/*

mv cdn.pocketcluster.io.conf /etc/nginx/conf.d/
mv nginx.conf /etc/nginx/
chown root:root /etc/nginx/conf.d/*
chown root:root /etc/nginx/nginx.conf
nginx -c /etc/nginx/nginx.conf -t
service nginx restart

# ---------------- CDN Serve Install ---------------- 
scp cdnserve.service us-west-01:~/
scp cdnserve us-west-01:~/

mv cdnserve /usr/local/bin/ && chown root:root /usr/local/bin/cdnserve
mv cdnserve.service /etc/systemd/system/ && chown root:root /etc/systemd/system/cdnserve.service

systemctl daemon-reload
systemctl start cdnserve
systemctl enable cdnserve
systemctl status cdnserve.service

# ---------------- Kernel Param Tuning ---------------- 
# show all parameters
$ sysctl -a

cat <<EOM >/etc/sysctl.conf
# Disble IPv6
net.ipv6.conf.all.disable_ipv6 = 1
net.ipv6.conf.default.disable_ipv6 = 1
net.ipv6.conf.lo.disable_ipv6 = 1

# Protection from SYN flood attack.
net.ipv4.tcp_syncookies = 1

# See evil packets in your logs.
net.ipv4.conf.all.log_martians = 1

# Increase number of incoming connections that can queue up before dropping
# The maximum number of "backlogged sockets".  Default is 128.
net.core.somaxconn = 2048

# Disable source routing and redirects
net.ipv4.conf.all.send_redirects = 0
net.ipv4.conf.all.accept_redirects = 0
net.ipv4.conf.all.accept_source_route = 0

# Disable packet forwarding.
net.ipv4.ip_forward = 0
net.ipv6.conf.all.forwarding = 0
EOM

# reload!
$ sysctl -p
# ---------------- Reboot ---------------- 


# -*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*- REPURPOSING CDN w/ CLOUDFLARE -*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
# - 1. Let's Encrypt -
mkdir -p /opt && cd /opt
git clone https://github.com/certbot/certbot /opt/letsencrypt

# Stop Cloudflare for the first time
service nginx stop

cd /opt/letsencrypt
./letsencrypt-auto certonly --standalone --email stkim1@colorfulglue.com -d us-west-01.cdn.pocketcluster.io

: <<OUTPUT
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Obtaining a new certificate
Performing the following challenges:
tls-sni-01 challenge for us-west-01.cdn.pocketcluster.io
Waiting for verification...
Cleaning up challenges

IMPORTANT NOTES:
 - Congratulations! Your certificate and chain have been saved at:
   /etc/letsencrypt/live/us-west-01.cdn.pocketcluster.io/fullchain.pem
   Your key file has been saved at:
   /etc/letsencrypt/live/us-west-01.cdn.pocketcluster.io/privkey.pem
   Your cert will expire on 2017-11-19. To obtain a new or tweaked
   version of this certificate in the future, simply run
   letsencrypt-auto again. To non-interactively renew *all* of your
   certificates, run "letsencrypt-auto renew"
 - If you like Certbot, please consider supporting our work by:

   Donating to ISRG / Let's Encrypt:   https://letsencrypt.org/donate
   Donating to EFF:                    https://eff.org/donate-le
OUTPUT

# - 2. NGINX(#1) CONFIG -
mv /etc/nginx/ssl/ /etc/nginx/ssl.crt/
rm /etc/nginx/conf.d/cdn.pocketcluster.org.conf

scp default us-west-01:~/
scp us-west-01.cdn.pocketcluster.io.conf us-west-01:~/
scp nginx.conf us-west-01:~/

mv default /etc/nginx/sites-available/
mv us-west-01.cdn.pocketcluster.io.conf /etc/nginx/conf.d/
mv nginx.conf /etc/nginx/

chown root:root /etc/nginx/sites-available/*
chown root:root /etc/nginx/conf.d/*
chown root:root /etc/nginx/nginx.conf

nginx -c /etc/nginx/nginx.conf -t
service nginx restart && service nginx status

# ---------------- home permission fix ---------------- 
sudo -s chown -R poimdall:pocketman ~/poimdall

# ---------------- cdn permission fix ---------------- 
sudo -s chown -R www-data:www-data /cdn-content
sudo -s chmod -R g+w /cdn-content
ls -lat /cdn-content

