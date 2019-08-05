package target

import (
	"fmt"
	"net"
	"strings"
)

type UnreachableTargetError struct {
	AddressesTried []net.IP
}

func (o UnreachableTargetError) Error() string {
	var at []string
	for _, i := range o.AddressesTried {
		at = append(at, i.String())
	}
	return fmt.Sprintf("cannot reach target using any of these addresses: %v", strings.Join(at, ", "))
}
