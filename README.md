# cisco-l2t

This is some tooling for working with the Cisco layer 2 traceroute service on Catalyst switches.

The layer 2 traceroute service is a UDP listener (port 2228) on (all?) Catalysts. It leaks information about configured VLANs, MAC address tables, interface names/speed/duplex, and CDP neighbors.

There is no requirement for L2 adjacency. The service is available via routed paths, provided that the querier can reach an L3 address on the target machine.

This library can create, send, receive and parse L2T messages. L2T messages allow you to enumerate configured VLANs, interrogate the L2 forwarding table, inquire about neighbors, interface names, speeds and duplex settings. All this without any credentials.

There's an example program for enumerating VLANs on a switch:

    $ ./enumerate-vlans 192.168.150.96
    86 VLANs found: 1-37 40-45 50-51 53-55 71-77 81-87 91-93 98-105 132 308-309 530 960-961 1100-1102 2000-2002 2222.
    $

Lots more examples (and a detailed readme) in the [cmd/lt2_ss directory](cmd/l2t_ss).

Truly mapping a Catalyst-based L2 topology will require some creativity because you can't interrogate the MAC population directly. Guessing well-known MAC addresses, swapping the src/dst query order (to elicit different error responses), noting STP root bridge address (and incrementing it on a per-vlan basis), etc... are all probably the kinds of things that a clever application using this library would be interested in doing.

The server implementation in switches is a little weird because it replies from the client-facing interface, not from the interface the client talked to. This means that the 5-tuple associated with the reply packet doesn't necessarily match the query packet. This makes it NAT un-friendly, and introduced a bunch of socket-layer challenges that this library attempts to handle as gracefully as possible. For stealthy network analysis, the client should run with a firewall configuration that avoids generating ICMP unreachables.
