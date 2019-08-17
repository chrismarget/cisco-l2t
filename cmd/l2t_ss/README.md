# cisco-l2t

This is some tooling for working with the Cisco layer 2 traceroute service on Catalyst switches.

It can create, send, receive and parse L2T messages. L2T messages allow you to enumerate configured VLANs, interrogate the L2 forwarding table, inquire about neighbors, interface names, speeds and duplex settings. All this without any credentials.

There's an example program for enumerating VLANs on a switch:

    poetaster:enumerate-vlans chris$ ./enumerate-vlans 192.168.96.150
    86 VLANs found: 1-37 40-45 50-51 53-55 71-77 81-87 91-93 98-105 132 308-309 530 960-961 1100-1102 2000-2002 2222.
    poetaster:enumerate-vlans chris$

Truly mapping a Catalyst-based L2 topology will require some creativity because you can't interrogate the MAC population directly. Guessing well-known MAC addresses, swapping the src/dst query order (to elicit different error responses), noting STP root bridge address (and incrementing it on a per-vlan basis), etc... are all probably the kinds of things that a clever application using this library would be interested in doing.
