package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/chrismarget/cisco-l2t/attribute"
	"github.com/chrismarget/cisco-l2t/communicate"
	"github.com/chrismarget/cisco-l2t/message"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	attrSep        = ":"
	typeFlag       = "t"
	verFlag        = "v"
	lenFlag        = "l"
	attrCountFlag  = "c"
	attrFlag       = "a"
	defaultVersion = 1
	printFlag      = "p"
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

func buildMsgBytes(t uint8, v uint8, l uint16, c uint8, asf attrStringFlags) ([]byte, error) {
	payload, err := getAttsBytesFromStrings(asf)
	if err != nil {
		log.Println(err)
		os.Exit(11)
	}

	bb := bytes.Buffer{}

	if flagProvided(typeFlag) {
		_, err = bb.Write([]byte{t})
	} else {
		_, err = bb.Write([]byte{uint8(message.RequestSrc)})
	}
	if err != nil {
		log.Println(err)
		os.Exit(12)
	}

	if flagProvided(verFlag) {
		_, err = bb.Write([]byte{v})
	} else {
		_, err = bb.Write([]byte{uint8(defaultVersion)})
	}
	if err != nil {
		log.Println(err)
		os.Exit(13)
	}

	b := make([]byte, 2)
	if flagProvided(lenFlag) {
		binary.BigEndian.PutUint16(b, l)
	} else {
		binary.BigEndian.PutUint16(b, uint16(len(payload)+5))
	}
	_, err = bb.Write(b)
	if err != nil {
		log.Println(err)
		os.Exit(14)
	}

	if flagProvided(attrCountFlag) {
		_, err = bb.Write([]byte{c})
	} else {
		_, err = bb.Write([]byte{uint8(len(asf))})
	}
	if err != nil {
		log.Println(err)
		os.Exit(15)
	}

	_, err = bb.Write(payload)
	if err != nil {
		log.Println(err)
		os.Exit(16)
	}

	return bb.Bytes(), nil
}

func main() {
	var attrStringFlags attrStringFlags
	flag.Var(&attrStringFlags, attrFlag, "attribute string form 'type:value' or raw TLV hex string")

	msgType := flag.Int(typeFlag, int(message.RequestSrc), "message type 0 - 255")
	msgVer := flag.Int(verFlag, int(message.Version1), "message version 0 - 255")
	msgLen := flag.Int(lenFlag, 0, "message length 0 - 65535")
	msgAC := flag.Int(attrCountFlag, 0, "message attribute count 0 255")
	doPrint := flag.Bool(printFlag, false, "attempt to parse/print outbound message")

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	payload, err := buildMsgBytes(
		uint8(*msgType),
		uint8(*msgVer),
		uint16(*msgLen),
		uint8(*msgAC),
		attrStringFlags,
	)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}

	if *doPrint {
		log.Println("do print")
		outMsg,err := message.UnmarshalMessageUnsafe(payload)
		if err != nil {
			log.Println(err)
			os.Exit(3)
		}

		fmt.Println(outMsg.String())
		for _, a := range outMsg.Attributes() {
			fmt.Println(a.String())
		}
	}

	//laddr := &net.UDPAddr{}
	raddr := &net.UDPAddr{
		IP:   net.ParseIP(flag.Arg(0)),
		Port: communicate.CiscoL2TPort,
	}

	//cxn, err := net.DialUDP(communicate.UdpProtocol, laddr, raddr )
	//if err != nil {
	//	log.Println(err)
	//	os.Exit(3)
	//}
	//
	//_, err = cxn.Write(payload)
	//if err != nil {
	//	log.Println(err)
	//	os.Exit(4)
	//}

	sendThis := communicate.SendThis{
		Payload:         payload,
		Destination:     raddr,
		ExpectReplyFrom: raddr.IP,
		RttGuess:        50 * time.Millisecond,
	}

	result := communicate.Communicate(sendThis, nil)
	if result.Err != nil {
		log.Println(result.Err)
		os.Exit(3)
	}

	inMsg, err := message.UnmarshalMessage(result.ReplyData)
	if err != nil {
		log.Println(result.Err)
		os.Exit(3)
	}

	log.Println(inMsg.String())
	for _, a := range inMsg.Attributes() {
		log.Println(a.String())
	}

	os.Exit(0)
}
