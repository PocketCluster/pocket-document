# ---------------- Setup Fail2Ban ---------------- 
bantime = 86400
destemail = root@localhost, stkim1@colorfulglue.com
sender = root@index.pocketcluster.org
port = 0:65535
banaction = iptables-multiport

scp jail.local index:~/config
scp filter.d/* index:~/config

(mv config/jail.local /etc/fail2ban/ || true) && (mv config/*.conf /etc/fail2ban/filter.d || true) && chown -R root:root /etc/fail2ban/*
touch /var/log/sshd.log 

service fail2ban restart

# force postfix to use ipv4 only
# http://www.postfix.org/IPV6_README.html
vi /etc/postfix/main.cf
inet_protocols = ipv4
service postfix restart

# ----------------- NGINX(#1) Config ---------------------
mkdir -p /etc/nginx/ssl.crt/
openssl dhparam -out /etc/nginx/ssl.crt/dhparams.pem 2048;

#mkdir -p /api-server && chown www-data:www-data /api-server && chmod 775 /api-server
#usermod -a -G www-data poimdall

scp default index:~/config
scp index.pocketcluster.io.conf index:~/config
scp nginx.conf index:~/config

mv config/default /etc/nginx/sites-available/
mv config/index.pocketcluster.io.conf /etc/nginx/conf.d/
mv config/nginx.conf /etc/nginx/

chown root:root /etc/nginx/sites-available/*
chown root:root /etc/nginx/conf.d/*
chown root:root /etc/nginx/nginx.conf

nginx -c /etc/nginx/nginx.conf -t
service nginx restart && service nginx status

# ------------ index service ------------
mv pocket-index.service /etc/systemd/system/ && chown root:root /etc/systemd/system/pocket-index.service

systemctl daemon-reload
systemctl start pocket-index
systemctl enable pocket-index
systemctl status pocket-index.service




# ----------------- NGINX(#1) Let's Encrypt Wildcard Domain Cert (10/22/2018) ---------------------

add-apt-repository ppa:certbot/certbot
apt-get update
apt-get install python-certbot-nginx

# revoke existing server cert
# https://letsencrypt.org/docs/revoking/
certbot revoke --cert-path /etc/letsencrypt/live/index.pocketcluster.io/fullchain.pem

# request wild card cert
certbot certonly \
--server https://acme-v02.api.letsencrypt.org/directory \
--email stkim1@colorfulglue.com \
-d pocketcluster.io \
-d *.pocketcluster.io \
--manual \
--preferred-challenges dns-01

: <<OUTPUT
> --server https://acme-v02.api.letsencrypt.org/directory \
> --email stkim1@colorfulglue.com \
> -d pocketcluster.io \
> -d *.pocketcluster.io \
> --manual \
> --preferred-challenges dns-01
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Plugins selected: Authenticator manual, Installer None
Starting new HTTPS connection (1): acme-v02.api.letsencrypt.org
Obtaining a new certificate
Performing the following challenges:
dns-01 challenge for pocketcluster.io
dns-01 challenge for pocketcluster.io

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
NOTE: The IP of this machine will be publicly logged as having requested this
certificate. If you're running certbot in manual mode on a machine that is not
your server, please ensure you're okay with that.

Are you OK with your IP being logged?
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
(Y)es/(N)o: Y

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Please deploy a DNS TXT record under the name
_acme-challenge.pocketcluster.io with the following value:

EMbaVoys_yALa80fgImFEwmtsUAho3YmtFMOjoW6lEg

Before continuing, verify the record is deployed.
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Press Enter to Continue

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Please deploy a DNS TXT record under the name
_acme-challenge.pocketcluster.io with the following value:

_XEyhe4xMAGXy3iNXl5FWl9mg-RbtSB20dZJaCbsr6s

Before continuing, verify the record is deployed.
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Press Enter to Continue
Waiting for verification...
Cleaning up challenges

IMPORTANT NOTES:
 - Congratulations! Your certificate and chain have been saved at:
   /etc/letsencrypt/live/pocketcluster.io/fullchain.pem
   Your key file has been saved at:
   /etc/letsencrypt/live/pocketcluster.io/privkey.pem
   Your cert will expire on 2019-01-20. To obtain a new or tweaked
   version of this certificate in the future, simply run certbot
   again. To non-interactively renew *all* of your certificates, run
   "certbot renew"
 - If you like Certbot, please consider supporting our work by:

   Donating to ISRG / Let's Encrypt:   https://letsencrypt.org/donate
   Donating to EFF:                    https://eff.org/donate-le
OUTPUT

# ----------------- NGINX(#1) Config ---------------------
mkdir -p /etc/nginx/ssl.crt/
openssl dhparam -out /etc/nginx/ssl.crt/dhparams.pem 2048;

mkdir -p /api-server && chown www-data:www-data /api-server && chmod 775 /api-server
usermod -a -G www-data birdman

scp default index:~/
scp index.pocketcluster.io.conf index:~/
scp nginx.conf index:~/

mv default /etc/nginx/sites-available/
mv index.pocketcluster.io.conf /etc/nginx/conf.d/
mv nginx.conf /etc/nginx/

chown root:root /etc/nginx/sites-available/*
chown root:root /etc/nginx/conf.d/*
chown root:root /etc/nginx/nginx.conf

nginx -c /etc/nginx/nginx.conf -t
service nginx restart && service nginx status


# ------------------ FIX NGINX PID NOT FOUND --------------
# https://www.digitalocean.com/community/questions/unable-to-start-nginx-failed-to-read-pid-from-file-run-nginx-pid
# https://bugs.launchpad.net/ubuntu/+source/nginx/+bug/1581864

mkdir /etc/systemd/system/nginx.service.d
printf "[Service]\nExecStartPost=/bin/sleep 0.1\n" > /etc/systemd/system/nginx.service.d/override.conf
systemctl daemon-reload
