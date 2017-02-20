# DHCP Enter / Exit Hook

isc-dhcp-client provides a way to trigger event notifications by placing hooks under `/etc/dhcp/dhclient-enter-hooks.d` and `/etc/dhcp/dhclient-exit-hooks.d`.

We'll be focusing on `/etc/dhcp/dhclient-exit-hooks.d` as it is the place to get renewed, bounded ip address status.

Executable hooks need to meet few conditions for execution

- Should be owned by root:root
- Must not have any extension including `.sh`

## Example

- Hook : [/etc/dhcp/dhclient-exit-hooks.d/variable_print](variable_print)
- Output : [/tmp/dhclient-script.debug](dhclient-script.debug)

## Reference

- [/etc/dhcp/](etc/dhcp/)
- [man dhclient-script](dhclient-script.man)
- [man dhclient.conf](dhclient.conf.man)
- [man dhclient.leases](dhclient.leases.man)
