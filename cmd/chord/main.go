package main

import (
	"flag"
	"fmt"
	"github.com/lukaspj/go-logging/logging"
	"github.com/lukaspj/go-chord/chord"
)

var logger = logging.GetLogger()

func main() {
	logger.SetLevel(logging.INFO)
	logger.AddStdoutOutput()

	port := flag.Int("sp", 5600, "Source port")
	host := flag.String("sh", "127.0.0.1", "Source host")
	id := flag.String("id", "", "id")
	dest := flag.String("dest", "", "Destination address")


	flag.Parse()

	var nid chord.NodeID
	if *id != "" {
		if len(*id) == 40 {
			// Assume hex value
			nid = chord.NewNodeIDFromString(*id)
		} else {
			nid = chord.NewNodeIDFromHash(*id)
		}
	} else {
		nid = chord.NewNodeIDFromHash(fmt.Sprintf("%s%d", *host, *port))
	}

	info := &chord.ContactInfo{
		Id: nid,
		Address: fmt.Sprintf("%s:%d", *host, *port),
	}

	peer := chord.NewPeer(info, *port)

	peer.Listen()

	if *dest != "" {
		peer.Connect(*dest)
	}

	<-make(chan struct{})
}
