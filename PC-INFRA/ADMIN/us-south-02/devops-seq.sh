# ---------------- FQDN ----------------
# us-south-02.cdn.pocketcluster.io

# ---------------- /etc/hosts ---------------- 
> vi /etc/hosts

#ff02::1        ip6-allnodes
#ff02::2        ip6-allrouters
 #::1           localhost ip6-localhost ip6-loopback

127.0.1.1       us-south-02.cdn.pocketcluster.io us-south-02 
127.0.0.1       localhost.localdomain localhost
# Auto-generated hostname. Please do not remove this comment.
198.46.233.209  us-south-02
::1             localhost

> service networking restart

# ---------------- check resolvconf ---------------- 
vi /etc/resolv.conf

search pocketcluster.io
nameserver 8.8.8.8
nameserver 8.8.4.4


# ---------------- Remove Packages ---------------- 
# apt list --installed > ~/deb-installed.log
# scp us-south-02.cdn.pocketcluster.io:~/deb-installed.log .

apt-get remove apache2* samba* libavahi* lynx* samba-common* bind9*
update-rc.d -f apache2 remove
update-rc.d -f bind9 remove
apt-get autoremove --purge && apt-get clean && apt-get autoclean && apt-get purge

# find /usr/share/locale -mindepth 1 -maxdepth 1 ! -name 'en*' | xargs rm -rf
# ---------------- Install Packages ---------------- 

apt update && apt upgrade -y && apt install --no-install-suggests -y dialog language-pack-en-base libpam-systemd dbus fail2ban logwatch iptables-persistent nginx git apparmor apparmor-profiles
apt-get install -y --no-install-recommends --no-install-suggests bind9-host

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

# ---------------- Create User ---------------- 
addgroup --gid 29999 pocketman &&\
adduser --gecos "PocketCluster (admin user)" --add_extra_groups --disabled-password --gid 29999 --uid 29999 poimdall &&\
usermod -a -G sudo -p $(mkpasswd -m sha-512 NONONPASSWD $(date +%m%H%M%S)) poimdall &&\
echo "poimdall ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/poimdall

# ssh-keygen -t rsa -b 4096 -C "stkim1@pocketcluster.io"
mkdir -p ~poimdall/.ssh && echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDRObM1vdIRjM0q7MoexT7ZY2tI/3wfzMwOcFf9A8RbIv3x6meUcmXkOUe33QheRT+w7vHevZBBO4yke4h4jFC3F1LL6UOIBAiSEZaJc/bVq2hK0UmD7EzMD0+VrrvRi/QgMS0e6lH+qp2tQRU6rBNjY0n1DuTYXh0JCck2phkuRwGpy8O7L9BLJfKZ/F2fKXLVdIjUan2xS+mmnp5QvPnXA5fWbCDFnBZY7Sor/Iq4CmEY41tpYtklOn4vfP8w3E/f+Yb0V9TcIkJzJxUsgtKUKU7cLvx8Tl0lUda2ql/JfP0hjBt48TnDAAF3pYF7FNb8V81FGit4/g60kzf+2MB662rru3Y2Md21qOwGsLmSSQeM2UuCycE+SWVzl47XpaZdem+MbNA7crhe85b7opyKylW2pKnWrlz6YUdpR788nMpON6oOBmHNYqVI1scvlP9qoighMGKoNT9fQb3qYEEmfhqd9RYrXi542XJ+JypX37rmZjOmM29ay2Z9VcRbgNrzyh2R6BywHABxPY/Uo5eNHBtAqq0L1YDAB417SFH6RHgD5invyVD5i5uS3S2OmkpIjyb6T3qgJiDG13H5K1+D8vJbLP14vnIeWuvsbADnzE6O+zBf/cv8L4iZhtN172wGcmRzjJU4h4hXWmUnD4WKcVbOzx0tqFZaH7qzUgFdNQ== stkim1@pocketcluster.io" > ~poimdall/.ssh/authorized_keys
chown -R poimdall:pocketman ~poimdall/.ssh && chmod 0700 ~poimdall/.ssh/ && chmod 0600 ~poimdall/.ssh/*

# ---------------- Setup SSHD ---------------- 
vi /etc/ssh/sshd_config

PermitRootLogin no
PermitEmptyPasswords no
PasswordAuthentication no
AuthorizedKeysFile      %h/.ssh/authorized_keys
AllowUsers poimdall

service ssh restart

# ---------------- Setup IpTables ---------------- 
mkdir -p /etc/iptables

scp rules.v4 us-south-02:~/
mv rules.v4 /etc/iptables/ && chown root:root /etc/iptables/*
iptables-apply /etc/iptables/rules.v4

iptables -S

# ---------------- Set Timezone ---------------- 
# https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
echo "US/Central" > /etc/timezone 
dpkg-reconfigure tzdata

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
dpkg-reconfigure locales

# ---------------- Setup Fail2Ban ---------------- 
bantime = 86400
destemail = root@localhost, stkim1@colorfulglue.com
sender = root@us-south-02.cdn.pocketcluster.io
port = 0:65535
banaction = iptables-multiport

scp jail.local us-south-02:~/
scp filter.d/* us-south-02:~/

touch /var/log/sshd.log 

(mv jail.local /etc/fail2ban/ || true) && (mv *.conf /etc/fail2ban/filter.d || true) && chown -R root:root /etc/fail2ban/*
service fail2ban restart

# force postfix to use ipv4 only 
# http://www.postfix.org/IPV6_README.html
vi /etc/postfix/main.cf
inet_protocols = ipv4
service postfix restart

# ---------------- Setup Logwatch ---------------- 
scp 00logwatch us-south-02:~/
mv 00logwatch /etc/cron.daily/ && chown root:root /etc/cron.daily/*

# ----------------- NGINX(#1) Let's Encrypt ---------------------
mkdir -p /opt && cd /opt
git clone https://github.com/certbot/certbot /opt/letsencrypt

# Stop Cloudflare for the first time
service nginx stop

cd /opt/letsencrypt
./letsencrypt-auto certonly --standalone --email stkim1@colorfulglue.com -d us-south-02.cdn.pocketcluster.io

: <<OUTPUT
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Obtaining a new certificate
Performing the following challenges:
tls-sni-01 challenge for us-south-02.cdn.pocketcluster.io
Waiting for verification...
Cleaning up challenges

IMPORTANT NOTES:
 - Congratulations! Your certificate and chain have been saved at:
   /etc/letsencrypt/live/us-south-02.cdn.pocketcluster.io/fullchain.pem
   Your key file has been saved at:
   /etc/letsencrypt/live/us-south-02.cdn.pocketcluster.io/privkey.pem
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

scp default us-south-02:~/
scp us-south-02.cdn.pocketcluster.io.conf us-south-02:~/
scp nginx.conf us-south-02:~/

mv default /etc/nginx/sites-available/
mv us-south-02.cdn.pocketcluster.io.conf /etc/nginx/conf.d/
mv nginx.conf /etc/nginx/

chown root:root /etc/nginx/sites-available/*
chown root:root /etc/nginx/conf.d/*
chown root:root /etc/nginx/nginx.conf

nginx -c /etc/nginx/nginx.conf -t
service nginx restart && service nginx status

# ---------------- CDN Serve Install ---------------- 
scp cdnserve.service us-south-02:~/
scp cdnserve us-south-02:~/

mkdir -p /cdn-content && chown www-data:www-data /cdn-content && chmod 775 /cdn-content
usermod -a -G www-data poimdall
usermod -aG adm poimdall

mv cdnserve /usr/local/bin/ && chown root:root /usr/local/bin/cdnserve
mv cdnserve.service /etc/systemd/system/ && chown root:root /etc/systemd/system/cdnserve.service

systemctl daemon-reload
systemctl start cdnserve
systemctl enable cdnserve
systemctl status cdnserve.service

# ---------------- Reboot ---------------- 
reboot

cat /proc/sys/net/ipv6/conf/all/disable_ipv6
1


# ---------------- home permission fix ---------------- 
sudo -s chown -R poimdall:pocketman ~/poimdall

# ---------------- cdn permission fix ---------------- 
sudo -s chown -R www-data:www-data /cdn-content
sudo -s chmod -R g+w /cdn-content
ls -lat /cdn-content

