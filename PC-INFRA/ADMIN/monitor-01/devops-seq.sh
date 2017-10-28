# ---------------- PARTITION SETUP ----------------
# ... check partition table
# Before you do this, make sure you know what what device you're dealing with
#sfdisk -d /dev/sdx > pine64-32gb-partition.layout
sfdisk /dev/sdx <pine64-32gb-partition.layout
resize2fs /dev/sdx2

# ---------------- FQDN ----------------
# monitor-01.internal.pocketcluster.io

# ---------------- /etc/hostname ---------------- 
echo "monitor-01" > /etc/hostname

# ---------------- /etc/hosts ---------------- 
127.0.0.1 localhost localhost
127.0.1.1 monitor-01.internal.pocketcluster.io monitor-01
192.168.1.154 monitor-01.internal.pocketcluster.io monitor-01
192.168.1.105 pc-master pc-master

# The following lines are desirable for IPv6 capable hosts
#::1 localhost ip6-localhost ip6-loopback
#fe00::0 ip6-localnet
#ff00::0 ip6-mcastprefix
#ff02::1 ip6-allnodes
#ff02::2 ip6-allrouters

# ---------------- Create User ---------------- 
addgroup --gid 29999 pocketman &&\
adduser --gecos "PocketCluster (admin user)" --add_extra_groups --disabled-password --gid 29999 --uid 29999 poimdall &&\
usermod -a -G sudo -p $(mkpasswd -m sha-512 pocket $(date +%m%H%M%S)) poimdall &&\
echo "poimdall ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/poimdall

#ssh-keygen -t rsa -b 4096 -C "stkim1@pocketcluster.io"
echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDRObM1vdIRjM0q7MoexT7ZY2tI/3wfzMwOcFf9A8RbIv3x6meUcmXkOUe33QheRT+w7vHevZBBO4yke4h4jFC3F1LL6UOIBAiSEZaJc/bVq2hK0UmD7EzMD0+VrrvRi/QgMS0e6lH+qp2tQRU6rBNjY0n1DuTYXh0JCck2phkuRwGpy8O7L9BLJfKZ/F2fKXLVdIjUan2xS+mmnp5QvPnXA5fWbCDFnBZY7Sor/Iq4CmEY41tpYtklOn4vfP8w3E/f+Yb0V9TcIkJzJxUsgtKUKU7cLvx8Tl0lUda2ql/JfP0hjBt48TnDAAF3pYF7FNb8V81FGit4/g60kzf+2MB662rru3Y2Md21qOwGsLmSSQeM2UuCycE+SWVzl47XpaZdem+MbNA7crhe85b7opyKylW2pKnWrlz6YUdpR788nMpON6oOBmHNYqVI1scvlP9qoighMGKoNT9fQb3qYEEmfhqd9RYrXi542XJ+JypX37rmZjOmM29ay2Z9VcRbgNrzyh2R6BywHABxPY/Uo5eNHBtAqq0L1YDAB417SFH6RHgD5invyVD5i5uS3S2OmkpIjyb6T3qgJiDG13H5K1+D8vJbLP14vnIeWuvsbADnzE6O+zBf/cv8L4iZhtN172wGcmRzjJU4h4hXWmUnD4WKcVbOzx0tqFZaH7qzUgFdNQ== stkim1@pocketcluster.io" > ~poimdall/.ssh/authorized_keys
chown -R poimdall:pocketman ~poimdall/.ssh && chmod 0700 ~poimdall/.ssh/ && chmod 0600 ~poimdall/.ssh/*

#usermod -p $(mkpasswd -m sha-512 pocket $(date +%m%H%M%S)) poimdall

# ---------------- Install Packages ---------------- 
apt update && apt upgrade -y && apt install --no-install-suggests -y dialog language-pack-en-base libpam-systemd dbus fail2ban logwatch iptables-persistent

# ---------------- Set Timezone ---------------- 
# https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
echo "Asia/Seoul" > /etc/timezone 
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
scp jail.local monitor-01:~/
mv jail.local /etc/fail2ban/ && chown root:root /etc/fail2ban/*
service fail2ban restart

# ---------------- Setup Logwatch ---------------- 
scp 00logwatch monitor-01:~/
mv 00logwatch /etc/cron.daily/ && chown root:root /etc/cron.daily/*

# ---------------- Setup IpTables ---------------- 
mkdir -p /etc/iptables
scp rules.v4 monitor-01:~/
mv rules.v4 /etc/iptables/ && chown root:root /etc/iptables/*
iptables-apply /etc/iptables/rules.v4

# ---------------- Kernel Tunning ---------------- 
# show all parameters
$ sysctl -a

scp sysctl.conf monitor-01:~/
mv sysctl.conf /etc/ && chown root:root /etc/sysctl.conf && sysctl -p

# ---------------- Remove Packages ---------------- 

apt list --installed

apt-get remove account-plugin-* cheese* deja-dup* libreoffice-* rhythmbox* speech-dispatcher* vlc* brasero* brltty* espeak* gdebi* hexchat* imagemagick* pidgin* shotwell* signon* simple-scan* speech-dispatcher* cups* printer-driver* transmission* synapse engrampa* plank* 

apt-get install mate-applets mate-dock-applet mate-indicator-applet mate-sensors-applet mate-gnome-main-menu-applet mailutils

deluser --remove-home ubuntu

# ---------------- Disable Avahi or remove it ---------------- 
# http://askubuntu.com/questions/867693/why-is-ubuntu-mate-using-google-dns-servers/868374
:<<DISABLE_AVAHI
After much investigation (with help from @Terrance) I discovered that Ubuntu Mate is using avahi-dnsconfd and some other device on my network (possibly a Raspberry Pi) was broadcasting Google's DNS servers over mDNS / Bonjor / Avahi.
avahi-dnsconfd was picking this up and when avahi-dnsconfd.action ran, it was calling resolvconf to add the DNS servers is discovered over mDNS to /etc/resolf.conf
DISABLE_AVAHI

```sh
sudo systemctl stop avahi-dnsconfd.service
sudo systemctl disable avahi-dnsconfd.service

pico /etc/default/avahi-daemon
AVAHI_DAEMON_DETECT_LOCAL=0 
```

or apt-get remove avahi-*

# ---------------- Permanent, static DNS ---------------- 
# http://askubuntu.com/questions/157154/how-do-i-include-lines-in-resolv-conf-that-wont-get-lost-on-reboot
# /etc/resolv.conf -> /run/resolvconf/resolv.conf. 
# Edit /etc/resolvconf/resolv.conf.d/base for permanent dns. Then, resolvconf -u
```
nameserver 192.168.1.1
nameserver 208.67.222.222
nameserver 208.67.220.220
nameserver 8.8.8.8
nameserver 8.8.4.4
```

# ---------------- Reboot ---------------- 
reboot

cat /proc/sys/net/ipv6/conf/all/disable_ipv6
1

```????
auto venet0:0
iface venet0:0 inet static
    address 107.181.138.177
    netmask 255.255.255.255
```

# ---------------- Install Uptime ----------------
apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 0C49F3730359A14518585931BC711F9BA15703C6
#echo "deb [ arch=amd64,arm64,ppc64el,s390x ] http://repo.mongodb.com/apt/ubuntu xenial/mongodb-enterprise/3.4 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-enterprise.list
echo "deb [ arch=arm64 ] http://repo.mongodb.com/apt/ubuntu xenial/mongodb-enterprise/3.4 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-enterprise.list
apt-get update && apt-get install -y mongodb-enterprise

systemctl enable mongod.service
service mongod start

# install distro-stable nodejs
apt-get install -y nodejs nodejs-legacy npm 
sudo -s mkdir -p /service && cd /service &&  git clone git://github.com/fzaninotto/uptime.git && chown -R poimdall:pocketman /service/uptime
scp ./default.yaml monitor-01:/service/uptime/config/ 

# setup mongo
$ mongo
> db.createCollection('uptime')
> use uptime
> db.createUser({
    user: "uptime",
    pwd: "uptime",
    roles: [{
        role: "readWrite",
        db: "uptime"
    }]
})
# if something goes wrong
> db.dropUser("uptime", {w: "majority", wtimeout: 5000})

$ nodejs app.js

# ---------------- CDN Configuration ---------------- 
openssl genrsa -out monitor-01.key 2048 -sha256
openssl req -new -key monitor-01.key -out monitor-01.csr -sha256 -subj "/CN=monitor-01" -config ./monitor-01.conf
openssl x509 -req -in monitor-01.csr -CA ca-cert.pub -CAkey ca-key.pem -CAcreateserial -out monitor-01.cert -sha256 -days 3650 -extensions v3_req -extfile monitor-01.conf
cat monitor-01.cert ca-cert.pub >> cert_chain.crt
cp monitor-01.key cert_chain.key

mkdir -p /cdn-content && chown www-data:www-data /cdn-content && chmod 775 /cdn-content
usermod -a -G www-data poimdall

scp default monitor-01:~/
scp monitor-01.conf monitor-01:~/
scp ./ssl/cert_chain.crt monitor-01:~/
scp ./ssl/cert_chain.key monitor-01:~/

mkdir -p /etc/nginx/ssl 
openssl dhparam -out /etc/nginx/ssl/dhparams.pem 2048
mv default /etc/nginx/sites-available/
chown root:root /etc/nginx/sites-available/*
mv cert_chain.key /etc/nginx/ssl/
mv cert_chain.crt /etc/nginx/ssl/
chown root:root /etc/nginx/ssl/*

mv monitor-01.conf /etc/nginx/conf.d/
chown root:root /etc/nginx/conf.d/*
nginx -c /etc/nginx/nginx.conf -t
service nginx reload
