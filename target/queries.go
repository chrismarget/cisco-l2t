package target

import (
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"net"
)

const (
	vlanMin = 1
	vlanMax = 4094
)

// HasIp returns a boolean indicating whether the target is known
// to have the given IP address.
func (o defaultTarget) HasIp(in *net.IP) bool {
	for _, i := range o.info {
		if in.Equal(i.destination.IP) {
			return true
		}
	}
	return false
}

// HasVlan queries the target about a VLAN is configured.
func (o *defaultTarget) HasVlan(vlan int) (bool, error) {
	var att attribute.Attribute
	var err error

	builder := message.NewMsgBuilder()
	builder.SetType(message.RequestSrc)
	att, err = attribute.NewAttrBuilder().SetType(attribute.SrcMacType).SetString("ffff.ffff.ffff").Build()
	if err != nil {
		return false, err
	}
	builder.SetAttr(att)

	att, err = attribute.NewAttrBuilder().SetType(attribute.DstMacType).SetString("ffff.ffff.ffff").Build()
	if err != nil {
		return false, err
	}
	builder.SetAttr(att)

	att, err = attribute.NewAttrBuilder().SetType(attribute.VlanType).SetInt(uint32(vlan)).Build()
	if err != nil {
		return false, err
	}
	builder.SetAttr(att)

	msg := builder.Build()
	err = msg.Validate()
	if err != nil {
		return false, err
	}

	response, err := o.Send(msg)
	if err != nil {
		return false, err
	}

	// Parse response. If we got the "good" error, then the VLAN exists.
	for _, a := range response.Attributes() {
		if a.Type() == attribute.ReplyStatusType &&
			a.String() == attribute.ReplyStatusSrcNotFound {
			return true, nil
		}
	}
	return false, nil
}

func (o *defaultTarget) GetIps() []net.IP {
	var out []net.IP
	INFO:
	for _, info := range o.info {
		ip := info.destination.IP
		for _, found := range out {
			if found.Equal(ip) {
				continue INFO
			}
		}
		out = append(out, ip)
	}
	return out
}

func (o *defaultTarget) GetVlans() ([]int, error) {
	var found []int
	for v := vlanMin; v <= vlanMax; v++ {
		vlanFound, err := o.HasVlan(v)
		if err != nil {
			return found, err
		}
		if vlanFound {
			found = append(found, v)
		}
	}
	return found, nil
}

func (o *defaultTarget) MacInVlan(mac net.HardwareAddr, vlan int) (bool, error) {
	if vlan < 1 || vlan > 4094 {
		return false, fmt.Errorf("vlan %d out of range", vlan)
	}

	var att attribute.Attribute
	var err error

	builder := message.NewMsgBuilder()
	builder.SetType(message.RequestSrc)
	att, err = attribute.NewAttrBuilder().
		SetType(attribute.SrcMacType).
		SetString("ffff.ffff.ffff").
		Build()
	if err != nil {
		return false, err
	}
	builder.SetAttr(att)

	att, err = attribute.NewAttrBuilder().
		SetType(attribute.DstMacType).
		SetString(mac.String()).
		Build()
	if err != nil {
		return false, err
	}
	builder.SetAttr(att)

	att, err = attribute.NewAttrBuilder().
		SetType(attribute.VlanType).
		SetInt(uint32(vlan)).
		Build()
	if err != nil {
		return false, err
	}
	builder.SetAttr(att)

	att, err = attribute.NewAttrBuilder().
		SetType(attribute.SrcIPv4Type).
		SetString(o.info[o.best].localAddr.String()).
		Build()
	if err != nil {
		return false, err
	}
	builder.SetAttr(att)

	msg := builder.Build()
	err = msg.Validate()
	if err != nil {
		return false, err
	}

	response, err := o.Send(msg)
	if err != nil {
		return false, err
	}

	if response.Attributes()[attribute.ReplyStatusType].Type() == 2 {
		return true, nil
	}

	return false, nil
}
