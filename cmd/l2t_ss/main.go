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
)

const (
	attrSep           = ":"
	defaultVersion    = 1
	typeFlag          = "t"
	typeFlagHelp      = "message type override: 0 - 255"
	verFlag           = "v"
	verFlagHelp       = "message version override: 0 - 255"
	lenFlag           = "l"
	lenFlagHelp       = "message length override: 0 - 65535 (default <calculated>)"
	attrCountFlag     = "c"
	attrCountFlagHelp = "message attribute count override: 0 255 (default <calculated>)"
	attrFlag          = "a"
	attrFlagHelp      = "attribute string form 'type:value' or raw TLV hex string"
	printFlag         = "p"
	printFlagHelp     = "attempt to parse/print outbound message (unsafe if sending broken messages)"
	usageTextCmd      = "[options] <catalyst-ip-address>"
	usageTextExplain  = "The following examples both create the same message:\n" +
		"  -a 2:0004.f284.dbbf -a 1:00:50:56:98:e2:12 -a 3:18 -a 14:192.168.1.2 <catalyst-ip-address>\n" +
		"  -t 2 -v 1 -l 31 -c 4 -a 02080004f284dbbf -a 010800505698e212 -a 03040012 -a 0e06c0a80102 <catalyst-ip-address>\n"
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
	flag.Var(&attrStringFlags, attrFlag, attrFlagHelp)
	msgType := flag.Int(typeFlag, int(message.RequestSrc), typeFlagHelp)
	msgVer := flag.Int(verFlag, int(message.Version1), verFlagHelp)
	msgLen := flag.Int(lenFlag, 0, lenFlagHelp)
	msgAC := flag.Int(attrCountFlag, 0, attrCountFlagHelp)
	doPrint := flag.Bool(printFlag, false, printFlagHelp)

	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(),
			"Usage :\n  %s %s\n%s\n\n",
			os.Args[0],
			usageTextCmd,
			usageTextExplain,
		)
		flag.PrintDefaults()
	}

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
		outMsg, err := message.UnmarshalMessageUnsafe(payload)
		if err != nil {
			log.Println(err)
			os.Exit(3)
		}

		fmt.Printf("Sending:  %s\n",outMsg.String())
		for _, a := range outMsg.Attributes() {
			fmt.Printf("  %2d %-20s %s\n",a.Type(),attribute.AttrTypeString[a.Type()],a.String())
		}
	}

	sendThis := communicate.SendThis{
		Payload: payload,
		Destination: &net.UDPAddr{
			IP:   net.ParseIP(flag.Arg(0)),
			Port: communicate.CiscoL2TPort,
		},
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

	fmt.Printf("Received: %s\n",inMsg.String())
	for _, a := range inMsg.Attributes() {
		fmt.Printf("  %2d %-20s %s\n",a.Type(),attribute.AttrTypeString[a.Type()],a.String())
	}

	os.Exit(0)
}
