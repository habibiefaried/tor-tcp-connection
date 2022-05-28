package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/cretz/bine/process/embedded"
	"github.com/cretz/bine/tor"
)

func main() {
	port := 5353
	fmt.Println("Starting and registering onion service, please wait a couple of minutes...")
	t, err := tor.Start(nil, &tor.StartConf{ProcessCreator: embedded.NewCreator(), DataDir: "data-dir-tor-to-tcp", EnableNetwork: true})
	if err != nil {
		log.Panicf("Unable to start Tor: %v", err)
	}
	defer t.Close()
	listenCtx, listenCancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer listenCancel()
	onion, err := t.Listen(listenCtx, &tor.ListenConf{Version3: true, RemotePorts: []int{5353}})
	if err != nil {
		log.Panicf("Unable to create onion service: %v", err)
	}
	defer onion.Close()

	fmt.Printf("Listening to %v.onion:%v\n", onion.ID, port)

	for {
		conn, err := onion.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}
		go func() {
			conn2, err := net.Dial("tcp", "localhost:8080")
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
