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

The header values are automatically calculated, but can be overwritten with the `-t`, `-l`, `-v` and `-c` options.

Attributes are added with the `-a` option, which takes one of two forms:

    -a <type>:<string-payload>
    -a hexstring

The `-a type:<string-payload>` form can be articulated with a type number (uint8), or a type name (see attribute/attr_common.go). String payloads are automatically formatted according to the requested type, and attribute lengths are calculated.

The `-a hexstring` form is correctly formatted when it includes type, length and payload.o

A `SrcMacType` payload containing MAC address FF:FF:FF:FF:FF:FF can be articulated as:

    -a 1:ffff.ffff.ffff
    -a SrcMacType:ff:ff:ff:ff:ff:ff
    -a 0108ffffffffffff

The `-p` option causes the program to attempt to print the outgoing message. This may or may not be possible, depending on whether the message is valid.
