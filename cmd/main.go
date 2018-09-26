package main

import (
	"log"
	"github.com/chrismarget/cisco-l2t/l2t"
)

func main() {
	log.Println(l2t.MakePortDuplex(0))
	log.Println(l2t.MakePortDuplex(1))
	log.Println(l2t.MakePortDuplex(2))
	log.Println(l2t.MakePortDuplex(3))
}
