# Node `pocketd` dns config guide

### `nameserver 127.0.0.1` blocks proper resolving

with `pocketd` on and `nameserver 127.0.0.1` entry in `resolv.conf`, system cannot properly resolve domain name. this need to be fixed.  
this involved with `resolvconf` conflicting with `isc-dhcp-cliet`.  

> Reference

- [nameserver 127.0.0.1](ping-trace-off.txt)
- [w/o nameserver 127.0.0.1](ping-trace-on.txt)

### `dhcpagent` is to receieve dhcp event.

1. [DEBUG] place a script in `/etc/dhcp/dhclient-exit-hooks.d/dhcpagent` with following content for debugging

  ```sh
  # Notifies DHCP event
  # Copyright 2017 PocketCluster.io

  echo "$(date): entering ${1%/*}, dumping variables." >> "/tmp/dh-client-env.log"
  /opt/pocket/bin/dhcpevent -mode=dhcpagent -dev=jsonprint | python -mjson.tool >>  "/tmp/dh-client-env.log"
  ```
  - to view the log, do the following

  ```sh
  #!/usr/bin/env bash

  echo "----------------------------------- show journal log ------------------------------------"
  /bin/journalctl -b0 --system _COMM=dhclient

  echo "-------------------------------- show noti exit hook log --------------------------------"
  cat /tmp/dh-client-env.log
  ```
2. [RELEASE] for production, remove all extra debugging info from `dhcpagent` with `0644` permission

  ```sh
  # Notifies DHCP event
  # Copyright 2017 PocketCluster.io

  /opt/pocket/bin/pocketd dhcpagent
  ```
  - you might add ` > /dev/null 2>&1` at the end.
  - <https://unix.stackexchange.com/questions/119648/redirecting-to-dev-null>

### `/etc/dhcp/dhclient.conf` 

(2017/11/01) As of now, `docker` & `pocketd` are separated and, as in, we cannot recompile `golang` to redirect name service. They need to share `nameserver` as `pocketd` provides name service.  

Moreover, we want the node to be able to use internal name service from dhcp server so that it can still have ways to at least get the name of host it needs.

We can go with `/etc/networking/interface` file's nameserver configuration, but that's rigid. So we'll go with `dhcpclient.conf` `prepend` option to add three dns servers ahead of local dns nameserver. 

```sh
vi /etc/dhcp/dhclient.conf

prepend domain-name-servers 127.0.0.1, 208.67.222.222, 208.67.220.220;
```

> References

- <https://askubuntu.com/questions/157154/how-do-i-include-lines-in-resolv-conf-that-wont-get-lost-on-reboot>
- <https://www.cyberciti.biz/faq/dhclient-etcresolvconf-hooks/>
- <https://use.opendns.com/>
- <https://linux.die.net/man/5/dhclient.conf>