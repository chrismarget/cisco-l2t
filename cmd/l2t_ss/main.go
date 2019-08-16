package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/message"
	"log"
	"net"
	"os"
	"strings"
)

const (
	attrSep        = ":"
	typeFlag       = "t"
	verFlag        = "v"
	lenFlag        = "l"
	attrCountFlag  = "c"
	attrFlag       = "a"
	defaultVersion = 1
	unsafeFlag     = "u"
)

type attrStringFlags []string

func (i *attrStringFlags) String() string {
	return "string representation of attrStringFlag"
}
func (i *attrStringFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func getAttFromStringOption(in string) (attribute.Attribute, error) {
	result := strings.SplitN(in, attrSep, 2)
	aType, err := attribute.AttrStringToType(result[0])
	if err != nil {
		return nil, err
	}

	att, err := attribute.NewAttrBuilder().SetType(aType).SetString(result[1]).Build()
	if err != nil {
		return nil, err
	}
	return att, nil
}

func getAttBytesFromHexStringOption(in string) ([]byte, error) {
	payload := make([]byte, hex.DecodedLen(len(in)))
	_, err := hex.Decode(payload, []byte(in))
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func getAttsBytesFromStrings(in []string) ([]byte, error) {
	var result []byte
	for _, s := range in {
		if strings.Contains(s, attrSep) {
			// attribute string option of the form type:value like:
			//  -a 3:101 (vlan 101)
			//  -a 2:0000.1111.2222 (destination mac 00:00:11:11:22:22)
			//  -a L2_ATTR_SRC_MAC:00:00:11:11:22:22 (source mac00:00:11:11:22:22)
			att, err := getAttFromStringOption(s)
			if err != nil {
				return result, err
			}
			result = append(result, byte(att.Type()))
			result = append(result, att.Len())
			result = append(result, att.Bytes()...)
		} else {
			// attribute string in hex payload TLV form like:
			//  -a 03040065 (vlan 101 where vlantype=3, len=4, value=0x0065 (101)
			b, err := getAttBytesFromHexStringOption(s)
			if err != nil {
				return result, err
			}
			result = append(result, b...)
		}
	}

	return result, nil
}

func flagProvided(name string) bool {
	var found bool
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func buildMsg(t uint8, v uint8, l uint16, c uint8, asf attrStringFlags) (message.Msg, error) {
	payload, err := getAttsBytesFromStrings(asf)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	bb := bytes.Buffer{}

	if flagProvided(typeFlag) {
		_, err = bb.Write([]byte{t})
	} else {
		_, err = bb.Write([]byte{uint8(message.RequestSrc)})
	}
	if err != nil {
		log.Println(err)
		os.Exit(3)
	}

	if flagProvided(verFlag) {
		_, err = bb.Write([]byte{v})
	} else {
		_, err = bb.Write([]byte{uint8(defaultVersion)})
	}
	if err != nil {
		log.Println(err)
		os.Exit(4)
	}

	if flagProvided(lenFlag) {
		_, err = bb.Write([]byte{uint8(l)})
	} else {
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(len(payload)+5))
		_, err = bb.Write(b)
	}
	if err != nil {
		log.Println(err)
		os.Exit(5)
	}

	if flagProvided(attrCountFlag) {
		_, err = bb.Write([]byte{c})
	} else {
		_, err = bb.Write([]byte{uint8(len(asf))})
	}
	if err != nil {
		log.Println(err)
		os.Exit(6)
	}

	_, err = bb.Write(payload)
	if err != nil {
		log.Println(err)
		os.Exit(7)
	}

	out, err := message.UnmarshalMessageUnsafe(bb.Bytes())
	if err != nil {
		log.Println(err)
		os.Exit(8)
	}

	return out, nil
}

func main() {
	var attrStringFlags attrStringFlags
	flag.Var(&attrStringFlags, attrFlag, "attribute string form 'type:value'")

	msgType := flag.Int(typeFlag, int(message.RequestSrc), "message type 0 - 255")
	msgVer := flag.Int(verFlag, int(message.Version1), "message version 0 - 255")
	msgLen := flag.Int(lenFlag, 0, "message length 0 - 65535")
	msgAC := flag.Int(attrCountFlag, 0, "message attribute count 0 255")
	unsafe := flag.Bool(unsafeFlag, false, "unsafe - don't attempt to parse outbound message")

	flag.Parse()

	outMsg, err := buildMsg(
		uint8(*msgType),
		uint8(*msgVer),
		uint16(*msgLen),
		uint8(*msgAC),
		attrStringFlags,
	)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if !*unsafe {
		fmt.Println(outMsg.String())
		for _, a := range outMsg.Attributes() {
			fmt.Println(a.String())
		}
	}

	destination := net.UDPAddr{
		IP:   nil,
		Port: 0,
		Zone: "",
	}
	cxn := net.DialUDP()

	os.Exit(0)
}
