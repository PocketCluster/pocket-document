# Docker Volume

### (11/02/2017)

**Docker Bridged Network + Volume Mouting need special care**

When recreating a service, `Unable to find a node that satisfies the following conditions : port 443 (Bridge mode)` could occur.  

> Reason & Workaround

1. From [issue #3469](https://github.com/docker/compose/issues/3469)
  
  It's happening because you have an anonymous volume in your container. You currently have two options:
  
  - Let the containers be recreated elsewhere, by mounting a named volume and using a volume driver which supports relocation of containers.
  - Destroy the old containers with docker-compose rm before creating the new ones.

2. From [issue 3420](https://github.com/docker/compose/issues/3420)
  
  The workaround is to first manually stop and remove the service and then run `docker-compose up -d` again.

3. From [Host ports and recreating containers](https://docs.docker.com/v1.12/compose/swarm/#/host-ports-and-recreating-containers)
  
  If a service maps a port from the host, e.g. 80:8000, then you may get an error like this when running docker-compose up on it after the first time:
  
  ```
  docker: Error response from daemon: unable to find a node that satisfies
  container==6ab2dfe36615ae786ef3fc35d641a260e3ea9663d6e69c5b70ce0ca6cb373c02.
  ```
  
  The usual cause of this error is that the container has a volume (defined either in its image or in the Compose file) without an explicit mapping, and so in order to preserve its data, Compose has directed Swarm to schedule the new container on the same node as the old container. This results in a port clash.
  
  There are two viable workarounds for this problem:
  
  - Specify a named volume, and use a volume driver which is capable of mounting the volume into the container regardless of what node it’s scheduled on. **_Compose does not give Swarm any specific scheduling instructions if a service uses only named volumes._**
  
  ```yaml
  version: "2"
  
  services:
    web:
      build: .
      ports:
        - "80:8000"
      volumes:
        - web-logs:/var/log/web
  
  volumes:
    web-logs:
      driver: custom-volume-driver
  ```
  
  - Remove the old container before creating the new one. You will lose any data in the volume.
  
  ```
  $ docker-compose stop web
  $ docker-compose rm -f web
  $ docker-compose up web
  ```


> Issues

- <https://github.com/docker/compose/issues/3420>
- <https://github.com/docker/compose/issues/3469>


**Named Volume Creation**

You can create named volume with driver and options specified (following issue [19990](https://github.com/moby/moby/issues/19990#issuecomment-248955005) `docker volume create --opt type=none --opt device=<host path> --opt o=bind`).

```yaml
services:
  namenode:
    image: pc-master:5000/x86_64/xenial:latest
    container_name: pc-core
    hostname: pc-core
    ports:
      - "8080:8080"
      - "8888:8888"
    networks:
      hadoop:
        ipv4_address: 172.16.128.1
    volumes:
      - data-hadoop:/pocket/hadoop
      - data-spark:/pocket/spark
    environment:
      - constraint:node==pc-core

volumes:
  data-hadoop: 
    driver: local
    driver_opts:
      type: none
      device: /pocket/hadoop/2.7.3
      o: bind
  data-spark: 
    driver: local
    driver_opts:
      type: none
      device: /pocket/spark/2.1.0
      o: bind
```

With named volume approach, 

1. We need to interact complicated volume driver-plugins and options. 
2. All the node would have volumes as well.
3. We need to clear exited process anyway.

> Issues

- **<https://stackoverflow.com/questions/35841241/docker-compose-named-mounted-volume>**
- <https://github.com/moby/moby/issues/19990>
- <https://github.com/docker/compose/issues/4188>
- <https://github.com/docker/compose/issues/4675>
- <https://stackoverflow.com/questions/36387032/how-to-set-a-path-on-host-for-a-named-volume-in-docker-compose-yml>
- <https://stackoverflow.com/questions/36312699/chown-docker-volumes-on-host-possibly-through-docker-compose>
- <https://serverfault.com/questions/761024/how-is-docker-compose-version-2-volumes-syntax-supposed-to-look>

> Reference

- **<https://docs.docker.com/v1.12/compose/compose-file/>**
- **<https://docs.docker.com/v1.12/engine/reference/commandline/volume_create/>**
- <https://docs.docker.com/compose/compose-file/compose-file-v2/#volume-configuration-reference>
- <https://docs.docker.com/compose/compose-file/compose-file-v2/#driver>
- <https://docs.docker.com/compose/compose-file/compose-versioning/>
- <https://docs.docker.com/engine/admin/volumes/volumes/>
- <https://docs.docker.com/engine/reference/commandline/volume_create/#examples>
- <https://docs.docker.com/engine/extend/legacy_plugins/#finding-a-plugin>
