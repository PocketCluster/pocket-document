# Comodo SSL

1. Create CSR

  ```sh
  openssl req -nodes -newkey rsa:2048 -keyout cdn.key -out cdn.csr -subj "/C=KR/ST=GyeongGi-Do/L=Seoul/O=PocketCluster/OU=Operation/CN=cdn.pocketcluster.org"
  ```
2. Copy [cdn.csr](cdn.csr) and submit to namecheap comodo
3. download `cdn_pocketcluster_org.zip` and unzip
4. combine host cert and ca cert.

  ```sh
  cat ./cdn_pocketcluster_org/cdn_pocketcluster_org.crt ./cdn_pocketcluster_org/cdn_pocketcluster_org.ca-bundle >> ../cdn_pocketcluster_org.crt
  ```
5. Copy key

  ```sh
  cp cdn.key ../cdn_pocketcluster_org.key
  ```