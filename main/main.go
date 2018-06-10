package main

import (
	"awesomeProject/src/github.com/lukaspj/go-chord"
	"flag"
	"fmt"
)

func main() {
	port := flag.Int("sp", 5600, "Source port")
	host := flag.String("sh", "127.0.0.1", "Source host")
	id := flag.String("id", "", "id")
	dest := flag.String("dest", "", "Destination address")


	flag.Parse()

	var nid go_chord.NodeID
	if *id != "" {
		if len(*id) == 40 {
			// Assume hex value
			nid = go_chord.NewNodeID(*id)
		} else {
			nid = go_chord.NewNodeIDFromHash(*id)
		}
	} else {
		nid = go_chord.NewRandomNodeID()
	}

	info := go_chord.ContactInfo{
		Id: nid,
		Address: fmt.Sprintf("%s:%d", *host, *port),
	}

	peer := go_chord.NewPeer(info, *port)

	peer.Listen()

	if *dest != "" {
		peer.Connect(*dest)
	}

	<-make(chan struct{})
}
