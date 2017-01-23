# Docker Daemon Configuration  
-

### Aug. 13, 2016  

- `/etc/default/docker`, put following options for debugging setup
  * `-H tcp://0.0.0.0:2375` : Network port
  * `-H unix:///var/run/docker.sock` : unix socket
  * `--insecure-registry 192.168.1.100:5000` : insecure repository without TLS  

- Docker Swarm Cluster Setup tutorials
  * [#SwarmWeek: Video Tutorials To Join Your First Docker Swarm Cluster](https://blog.docker.com/2016/03/swarmweek-join-your-first-swarm/)
  * Docker Daemon <https://docs.docker.com/v1.10/engine/reference/commandline/daemon/>
  * Authorization Plugin <https://docs.docker.com/v1.10/engine/extend/authorization/>
  * Protect the Docker daemon socket <https://docs.docker.com/engine/security/https/>
