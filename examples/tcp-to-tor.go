package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/cretz/bine/tor"
)

func main() {
	fmt.Println("Starting and registering onion service, please wait a couple of minutes...")
	t, err := tor.Start(nil, &tor.StartConf{DataDir: "data-dir-tcp-to-tor", EnableNetwork: true})
	if err != nil {
		log.Panicf("Unable to start Tor: %v", err)
	}

	defer t.Close()
	// Wait at most a minute to start network and get
	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
	defer dialCancel()

	dialer, err := t.Dialer(dialCtx, nil)
	if err != nil {
		log.Panicf("error dialing remote addr", err)
	}

	fmt.Println("Listening...")
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("error accepting connection", err)
			continue
		}

		go func() {
			fmt.Println("Connecting...")
			conn2, err := dialer.Dial("tcp", "i4owyifwx2xctzfblsri45ox5uvkcawt3owvbx5wk4hczcyemae6yhad.onion:5353")
			if err != nil {
				log.Println("error dialing remote addr", err)
				return
			}

			go io.Copy(conn2, conn)
			io.Copy(conn, conn2)
			conn2.Close()
			conn.Close()
		}()
	}
}
