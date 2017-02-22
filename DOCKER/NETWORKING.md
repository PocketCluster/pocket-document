# Docker Networking

### (08/14/2016)

**Linux Fan Networking is not required for Docker Overlay Network**

- What is FAN networking?
  >  Essentially, what you need is an address multiplier – anywhere you have one interface, it would be handy to have 250 of them instead.

  > So we came up with the “fan”. It’s called that because you can picture it as a fan behind each of your existing IP addresses, with another 250 IP addresses available. Anywhere you have an IP you can make a fan, and every fan gives you 250x the IP addresses. More than that, you can run multiple fans, so each IP address could stand in front of thousands of container IP addresses.

- Since docker has its own network tunnelling solution, Ubuntu FAN is not required.
  > VXLAN encapsulation to build an overlay network of “tunnels” between the Docker hosts that would allow the container tra c to pass unmodi ed over the physical network.

> References

- <https://insights.ubuntu.com/2015/06/24/introducing-the-fan-simpler-container-networking/>
- <https://wiki.ubuntu.com/FanNetworking>
- <https://www.singlestoneconsulting.com/~/media/files/whitepapers/dockernetworking2.pdf>

