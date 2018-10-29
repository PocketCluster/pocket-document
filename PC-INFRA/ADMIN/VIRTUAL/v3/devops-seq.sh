scp * index:~/

---

mv default /etc/nginx/sites-available/
mv api.pocketcluster.io.conf /etc/nginx/conf.d/
mv index.pocketcluster.io.conf /etc/nginx/conf.d/
mv news.pocketcluster.io.conf /etc/nginx/conf.d/
mv nginx.conf /etc/nginx/
rm devops-seq.sh

chown root:root /etc/nginx/sites-available/*
chown root:root /etc/nginx/conf.d/*
chown root:root /etc/nginx/nginx.conf

nginx -c /etc/nginx/nginx.conf -t

service nginx restart && service nginx status

---
curl --head index.pocketcluster.io
curl --head news.pocketcluster.io
curl -iv --head api.pocketcluster.io
curl -iv --head pocketcluster.io

curl -iv https://api.pocketcluster.io/service/v014/package/list
curl -iv https://api.pocketcluster.io/healthcheck

