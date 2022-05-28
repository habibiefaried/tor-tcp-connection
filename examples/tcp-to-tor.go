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
	onionDNSaddress := "pdnshs76a5djmxzb4c5chvzc5kpv6egk3htorh7u4okvcpdcdj5r5bqd.onion:53"
	fmt.Println("Starting and registering onion service, please wait a couple of minutes...")
	t, err := tor.Start(nil, &tor.StartConf{ProcessCreator: embedded.NewCreator(), DataDir: "data-dir-tcp-to-tor", EnableNetwork: true, DebugWriter: nil})
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

	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		panic(err)
	}

	// For testing the DNS resolve via TOR network
	fmt.Println("Testing and building network...")
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {

			return dialer.Dial("tcp", onionDNSaddress)
		},
	}
	ip, _ := r.LookupHost(context.Background(), "puredns.org")

	fmt.Println("IP resolv: " + ip[0])
	fmt.Println("Listening...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("error accepting connection", err)
			continue
		}

		go func() {
			fmt.Println("Connecting...")
			conn2, err := dialer.Dial("tcp", onionDNSaddress)
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
