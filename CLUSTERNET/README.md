# Multihost Container Cluster Networking Setup

Multihost Container Cluster Networoking could be achieved with few options. We'll cover `overlay` type networking here.

## Requirements

Here's precondition of what's necessary

- Each involved hosts kernel to support `CONFIG_VXLAN` option. 
- Swarm does not need to connect to a backend K/V service. It can work with `node://` variable  
  * Swarm can be connected backend K/V such as Consul or ETCD for discovery service  
- Backend K/V service such as ETCD/Consul must be available and properly setup. See [ETCD](../ETCD/README.md)
- In case `TLS` is to be supported, then each hosts should come with their own `cert` and `key`.
  * This implies any service that has different IP should have their own cert-key
  * It is recommneded to have a separate cert-key pair for each service to be secure  

## Docker Daemon Option

Docker Daemon must be running with following options.

- `--cluster-advertise=eth0:2375` cluster advertise should be done on the primary network interface
- `--cluster-store=etcd://192.168.1.150:2379` cluster storage to point K/V service ip and port
- Docker Daemon option should look like following
  
  `/etc/default/docker`
  
  ```sh
  DOCKER_OPTS="--tlsverify --tlscacert=/etc/docker/ca-cert.pub --tlscert=/etc/docker/pocket.cert --tlskey=/etc/docker/pocket-key.pem -H tcp://0.0.0.0:2375 -H unix:///var/run/docker.sock --cluster-advertise=eth0:2375 --cluster-store=etcd://192.168.1.150:2379"
  ```
- If containers need to bind only to the local network, then use `-H 192.168.10.0/24` instead
- Repeat this setup on every node

## ETCD Service

See [here](../ETCD/README.md)

## Swarm Manager 

- Swarm can run without K/V backend using `node://`

  ```sh
  swarm --debug manage --tlsverify=true --tlscacert=ca-cert.pub --tlscert=swarm.cert --tlskey=swarm.key --host=:3376 --advertise=192.168.1.105:3376 nodes://192.168.1.150:2375,192.168.1.151:2375,192.168.1.152:2375,192.168.1.153:2375,192.168.1.154:2375
  ```
- <sup>*</sup>You can use ETCD as service discovery with `-discovery-opt` option to specify K/V path
  
  ```sh
  swarm manage -H :3375 --replication --advertise 192.168.10.1:3375 --discovery-opt kv.path=docker/nodes etcd://192.168.10.1:2379
  ```

## Create Docker Network
Create Docker network. You can do this on individual docker host or use `docker-compose`. 

- Ex #1
  
  ```sh
  docker network create --driver=bridge --subnet=172.28.5.0/24 --ip-range=172.28.5.0/24 --gateway=172.28.5.254 hadoop
  ```
- Ex #2

  ```sh
  docker network create -d overlay --subnet=192.170.0.0/16 --gateway=192.170.0.100 --aux-address a=192.170.1.5 --aux-address b=192.170.1.6 my-multihost-network
  ```

## Docker Compose for multi-host network

Launch with `docker-compose.yml`. See [this example](docker-compose.yml)

`192.168.1.105:3376` is Swarm host

```sh
docker-compose --host 192.168.1.105:3376 --tlsverify --tlscacert ca-cert.pub --tlscert composer.cert --tlskey composer.key up
```

> Reference

- [Docker container networking](https://docs.docker.com/engine/userguide/networking/)
- [Get started with multi-host networking](https://docs.docker.com/engine/userguide/networking/get-started-overlay/)
- [Why Compose creates network in modular fashion](https://github.com/docker/docker/issues/20547)
- <http://stackoverflow.com/questions/34892377/the-cluster-store-and-cluster-advertise-dont-work>