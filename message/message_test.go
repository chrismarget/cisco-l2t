package message

import (
	"github.com/chrismarget/cisco-l2t/attribute"
	"log"
	"testing"
)

func TestBytesToAttrSlice(t *testing.T) {
	var result []attribute.Attr
	var err error

	result, err = bytesToAttrSlice([]byte{11, 3, 1})
	if err != nil {
		t.Error(err)
	}
	log.Println(result)

	//result, err = bytesToAttrSlice([]byte{1, 8, 0,0,0,0,0,0})
	//if err != nil {
	//	t.Error(err)
	//}
	//log.Println(result)
	//
	//result, err = bytesToAttrSlice([]byte{1, 9, 0,0,0,0,0,0,0})
	//if err != nil {
	//	t.Error(err)
	//}
	//log.Println(result)

}
