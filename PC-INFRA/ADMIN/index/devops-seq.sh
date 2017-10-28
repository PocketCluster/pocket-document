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
