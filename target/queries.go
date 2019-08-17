package target

import (
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
)

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

	att, err = attribute.NewAttrBuilder().SetType(attribute.VlanType).SetInt(uint32(i)).Build()
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
