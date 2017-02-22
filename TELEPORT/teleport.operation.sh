teleport start -d --roles=node --token=d52527f9-b260-41d0-bb5a-e23b0cfe0f8f --auth-server=192.168.1.150
teleport start -d --roles=node --token=c9s93fd9-3333-91d3-9999-c9s93fd98f43 --auth-server=192.168.1.150

> tctl users add root root
https://192.168.1.150:3080/web/newuser/5a563318cb0b2b001f1a4e502d1d965d

> tsh --proxy=192.168.1.150 ssh --user=root root@192.168.1.151
> tsh --insecure --proxy=192.168.1.150 ssh --user=root root@192.168.1.151
> tsh --insecure --proxy=192.168.1.150 ssh --user=root root@192.168.1.151 '/usr/sbin/service --status-all'
WARNING: You are using insecure connection to SSH proxy https://192.168.1.150:3080
Enter password for Teleport user root:
Enter your HOTP token:
642578


// for pc-core connection
> tsh --insecure --proxy=192.168.1.248 ssh --user=root root@192.168.1.151
WARNING: You are using insecure connection to SSH proxy https://192.168.1.248:3080
dial tcp 192.168.1.248:3080: getsockopt: connection refused


> /opt/gopkg/src/github.com/stkim1/teleclient/main.go
ERRO[0000] could not find matching keys for 192.168.1.248:3023 at 192.168.1.248:3023
WARNING: You are using insecure connection to SSH proxy https://192.168.1.248:3080
dial tcp 192.168.1.248:3080: getsockopt: connection refused
exit status 1