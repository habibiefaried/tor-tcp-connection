package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/cretz/bine/tor"
)

func gracefulExit(err error) {
	fmt.Println("Error has occured...")
	fmt.Println(err)
	fmt.Print("press Enter to exit")
	var b byte
	_, _ = fmt.Scanf("%v", &b)
}

func main() {
	onionDNSaddress := "pdnshs76a5djmxzb4c5chvzc5kpv6egk3htorh7u4okvcpdcdj5r5bqd.onion:53"
	fmt.Println("Starting and registering onion service, please wait a couple of minutes...")
	var t *tor.Tor
	var err error
	if runtime.GOOS == "windows" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Full path of Tor Browser Installation: ")
		text, _ := reader.ReadString('\n')
		if err != nil {
			gracefulExit(err)
		}
		torPath := fmt.Sprintf("%v\\Browser\\TorBrowser\\Tor\\tor.exe", strings.TrimSpace(text))
		fmt.Println("Opening tor.exe at " + torPath)
		t, err = tor.Start(nil, &tor.StartConf{ExePath: torPath, DataDir: "data-dir-tcp-to-tor", EnableNetwork: true, DebugWriter: nil})
		if err != nil {
			gracefulExit(err)
		}
	} else {
		t, err = tor.Start(nil, &tor.StartConf{DataDir: "data-dir-tcp-to-tor", EnableNetwork: true, DebugWriter: nil})
		if err != nil {
			gracefulExit(err)
		}
	}

	defer t.Close()
	// Wait at most a minute to start network and get
	dialCtx, dialCancel := context.WithTimeout(context.Background(), time.Minute)
	defer dialCancel()

	dialer, err := t.Dialer(dialCtx, nil)
	if err != nil {
		gracefulExit(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:53")
	if err != nil {
		fmt.Println("Please free-up your port 53 before using this")
		gracefulExit(err)
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
