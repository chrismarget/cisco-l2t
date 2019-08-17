# l2t_ss

Layer 2 Traceroute Simple Sender

This program sends L2T messages, listens for responses, prints the responses.

It doesn't use many of the attribute and builder methods available in this library because they're validated and this sender is deliberately unsafe. It can generate bogus messages, might be useful for fuzzing the L2T service on a switch.

The minimal valid CLI arguments sends a correctly formatted (but invalid) message to the target switch:

    l2t_ss <target-ip>

This message consists of a header only and looks like: 

    0x 02 01 00 05 00
       -- -- ----- --
       |  |  |     +---> Attribute Count: 0
       |  |  |
       |  |  +---------> Total Message Length: 5 bytes
       |  |
       |  +------------> Message Version: 1
       |
       +---------------> Message Type: 2 (L2T_REQUEST_SRC)

The header values are automatically calculated, but can each be overwritten with the `-t`, `-l`, `-v` and `-c` options.

Attributes are added with the `-a` option, which takes one of two forms:

    -a <type>:<string-payload>
    -a hexstring

The `-a type:<string-payload>` form can be articulated with a type number (uint8), or a type name (see attribute/attr_common.go). String payloads are automatically formatted according to the requested type, and attribute lengths are calculated.

The `-a hexstring` form is correctly formatted when it includes type, length and payload.

A `SrcMacType` payload containing MAC address FF:FF:FF:FF:FF:FF can be articulated as:

    -a 1:ffff.ffff.ffff
    -a SrcMacType:ff:ff:ff:ff:ff:ff
    -a 0108ffffffffffff

Attributes are packed into outgoing messages in the order they're provided on the CLI.

The `-p` option causes the program to attempt to print the outgoing message. This may or may not be possible, depending on whether the message is valid.

##Some complete examples:

Determine whether VLAN 100 exists on a switch:

    $ ./l2t_ss -p -t 2 -a 1:ffff.ffff.ffff -a 2:ffff.ffff.ffff -a 3:100 -a 14:<local-ip-address> <target-switch-ip-address>
    Sending:  L2T_REQUEST_SRC (2) with 4 attributes (31 bytes)
       1 L2_ATTR_SRC_MAC      ff:ff:ff:ff:ff:ff
       2 L2_ATTR_DST_MAC      ff:ff:ff:ff:ff:ff
       3 L2_ATTR_VLAN         100
      14 L2_ATTR_SRC_IP       <local-ip-address>
    Received: L2T_REPLY_SRC (4) with 4 attributes (34 bytes)
       4 L2_ATTR_DEV_NAME     <switch-hostname-string>          <- switch reveals its hostname
       5 L2_ATTR_DEV_TYPE     <switch-platform-string>          <- switch reveals its platform type
       6 L2_ATTR_DEV_IP       <switch-ip-address>               <- this might be a new IP address we didn't know about
      15 L2_ATTR_REPLY_STATUS Source Mac address not found      <- VLAN 100 exists (would be 'Status unknown (3)' otherwise)

Determine whether MAC address 0011.2233.4455 exists within VLAN 100:

    $ ./l2t_ss chris$ ./l2t_ss -t 2 -a 1:ffff.ffff.ffff -a 2:0011.2233.4455 -a 3:100 -a <local-ip-address> <target-switch-ip-address>
    Received: L2T_REPLY_SRC (4) with 6 attributes (62 bytes)
       4 L2_ATTR_DEV_NAME     <switch-hostname-string>          <- switch reveals its hostname
       5 L2_ATTR_DEV_TYPE     <switch-platform-string>          <- switch reveals its platform type
       6 L2_ATTR_DEV_IP       <switch-ip-address>               <- this might be a new IP address we didn't know about
      13 L2_ATTR_NBR_IP       <neighbor-cdp-mgmt-ip>            <- switch reveals CDP management IP of neighbor switch
      15 L2_ATTR_REPLY_STATUS Status unknown (2)                <- MAC exists (would be 'Source Mac address not found' otherwise)
      16 L2_ATTR_NBR_DEV_ID   <neighbor-switch-name>            <- switch reveals host/domain name of neighbor switch

Queries of type 1 reveal even more:

	$ ./l2t_ss chris$ ./l2t_ss -t 1 -a 1:<mac-string> -a 2:<mac-string-again> -a 3:100 -a 14:0.0.0.0 <target-switch-ip-address>
	Received: L2T_REPLY_DST (3) with 12 attributes (96 bytes)
   	4 L2_ATTR_DEV_NAME     <switch-hostname-string>         <- switch reveals its hostname
   	5 L2_ATTR_DEV_TYPE     <switch-platform-string>         <- switch reveals its platform type
   	6 L2_ATTR_DEV_IP       <switch-ip-address>              <- this might be a new IP address we didn't know about
   	7 L2_ATTR_INPORT_NAME  Gi1/2                            <- interface facing toward attr_type 1 MAC address
   	8 L2_ATTR_OUTPORT_NAME Gi1/2                            <- interface facing toward attr_type 2 MAC address
   	9 L2_ATTR_INPORT_SPEED 1Gb/s                            <- this is crazy, right?
  	10 L2_ATTR_OUTPORT_SPEED 1Gb/s                          <- these values don't render if speeds are "auto"
  	11 L2_ATTR_INPORT_DUPLEX Auto                           <- duplex command wasn't typed in the config
  	12 L2_ATTR_OUTPORT_DUPLEX Auto                          <- otherwise we'd see 'full' or 'half'
  	13 L2_ATTR_NBR_IP       192.168.56.12                   <- neighbor switch IP - we should poke this one next!
  	15 L2_ATTR_REPLY_STATUS Status unknown (2)              <- still working on all the reply status codes
  	16 L2_ATTR_NBR_DEV_ID   switch3.company.com             <- the neighbor's hostname
