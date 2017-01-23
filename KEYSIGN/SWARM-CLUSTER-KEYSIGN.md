## How to sign keys with certificate

### Create CA Key / public CA Authority

1. Create a private key called `ca-key.pem` for the CA
	
  > openssl genrsa -out ca-key.pem 2048 -sha256
  
2. Inspect the private key

  > openssl rsa -in ca-key.pem -noout -text

3. Create a public CA Authority called `ca-cert.pub` for the CA

  > openssl req -x509 -new -nodes -key ca-key.pem -out ca-cert.pub -sha256 -days 365 -subj '/CN=pc-ca'

4. Inspect the public authority

  > openssl x509 -in ca-cert.pub -noout -text

### Create Master Key/ Certificate Sign Request/ Signed Certificate

1. Create a private key called `master-key.pem` for the CA
	
  > openssl genrsa -out master-key.pem 2048 -sha256
  
2. Inspect the private key

  > openssl rsa -in master-key.pem -noout -text

3. Create sign request called `master.csr`

  > openssl req -new -key master-key.pem -out master.csr -sha256 -subj "/CN=pc-master" -config master.conf
  
4. Inspect the Certificate Sign Request

  > openssl req -in master.csr -noout -text

5. Create signed certificate called `master.cert`

  > openssl x509 -req -in master.csr -CA ca-cert.pub -CAkey ca-key.pem -CAcreateserial -out master.cert -sha256 -days 365 -extensions v3_req -extfile master.conf

8. Inspect the created self CA certificate

  > openssl x509 -in master.cert -noout -text

- `master.conf` 

  ```sh
  [req]
  req_extensions 				= v3_req
  distinguished_name 			= req_distinguished_name
  
  [req_distinguished_name]
  countryName					= KR
  commonName					= pc-master

  [ v3_req ]
  basicConstraints 				= CA:FALSE
  keyUsage 						= nonRepudiation, digitalSignature, keyEncipherment
  extendedKeyUsage 				= serverAuth, clientAuth
  ```

### Create Node Key/ Certificate Sign Request/ Signed Certificate

1. Create a private key `node-key.pem` for your Swarm Manager:

  > openssl genrsa -out node-key.pem 2048 -sha256

2. Generate a **certificate signing request** (CSR) `node.csr` using the private key you create in the previous step:

  > openssl req -new -key node-key.pem -out node.csr -sha256 -subj "/CN=pc-node1" -config node.conf

3. Create the certificate `node.cert` based on the CSR created in the previous step.

  > openssl x509 -req -in node.csr -CA ca-cert.pub -CAkey ca-key.pem -CAcreateserial -out node.cert -sha256 -days 365 -extensions v3_req -extfile node.conf

- `node.conf`

  ```sh
  [req]
  req_extensions 				= v3_req
  distinguished_name 			= req_distinguished_name
  
  [req_distinguished_name]
  countryName					= KR
  commonName					= pc-node1
  
  [ v3_req ]
  basicConstraints 				= CA:FALSE
  keyUsage 						= nonRepudiation, digitalSignature, keyEncipherment
  extendedKeyUsage 				= serverAuth, clientAuth
  subjectAltName 				= @alt_names
  
  [alt_names]
  DNS.1 						= pocketcluster.local
  IP.1 							= 192.168.1.152
  ```

### Docker Key/Cert Gen Script
- run [`docker-keycert`](docker-keycert/dockercertgen.sh) with following options
  
  ```sh
  $ dockercertgen.sh <host_name> <ip_address>
  ```

### Docker Daemon Config
- Copy key/cert/pubkey files to `node` (e.g. with `rsync`, `scp`)
- Move files to an appropriate path (e.g. `/etc/docker`) and change permission

  ```sh
  chmod -v 0400 ca-cert.pub node.cert node-key.pem
  ```
- Edit `/lib/systemd/system/docker.service` or `/etc/default/docker` and insert following options
  
  ```sh
  DOCKER_OPTS="--dns=172.17.42.1  --tlsverify --tlscacert=/etc/docker/ca-cert.pub --tlscert=/etc/docker/odroid.cert --tlskey=/etc/docker/odroid-key.pem -H tcp://0.0.0.0:2375 -H unix:///var/run/docker.sock --insecure-registry pc-master:5000"

  # optional
  DOCKER_OPTS="--dns=172.17.42.1 --insecure-registry pc-master:5000"
  ```
- If you want docker to listen from a single host change `-H tcp://0.0.0.0:2375` to `-H=<HOST>:2376`

> References

- [dockerd [OPTIONS]](https://docs.docker.com/engine/reference/commandline/dockerd/)
- [Protect the Docker daemon socket](https://docs.docker.com/engine/security/https/)
- [Configure and run Docker on various distributions](https://docs.docker.com/engine/admin/)

### Swarm Manager Command sequence

- This is a shortened version of the above sequence

  ```sh
  openssl genrsa -out ca-key.pem 2048 -sha256
  
  openssl req -x509 -new -nodes -key ca-key.pem -out ca-cert.pub -sha256 -days 365 -subj '/CN=pc-ca'
  
  >>>>
  
  openssl genrsa -out master-key.pem 2048 -sha256
  
  openssl req -new -key master-key.pem -out master.csr -sha256 -subj "/CN=pc-master" -config master.conf
  
  openssl x509 -req -in master.csr -CA ca-cert.pub -CAkey ca-key.pem -CAcreateserial -out master.cert -sha256 -days 365 -extensions v3_req -extfile master.conf
  
  >>>>
  
  openssl genrsa -out node-key.pem 2048 -sha256
  
  openssl req -new -key node-key.pem -out node.csr -sha256 -subj "/CN=pc-node1" -config node.conf
  
  openssl x509 -req -in node.csr -CA ca-cert.pub -CAkey ca-key.pem -CAcreateserial -out node.cert -sha256 -days 365 -extensions v3_req -extfile node.conf
  
  >>>>>
  
  # use nodes:// to connect multiple node
  swarm --debug manage --tlsverify=true --tlscacert=ca-cert.pub --tlscert=master.cert --tlskey=master-key.pem --host=:3376 --advertise=192.168.1.236:3376 nodes://192.168.1.151:2375,192.168.1.152:2375
  ```

> References

- [Use Docker Swarm with TLS
](https://docs.docker.com/swarm/secure-swarm-tls/)
- [Configure Docker Swarm for TLS](https://docs.docker.com/swarm/configure-tls/)
- [Plan for Swarm in production](https://docs.docker.com/swarm/plan-for-production/)
- [Build a Swarm cluster for production
](https://docs.docker.com/swarm/install-manual/)
