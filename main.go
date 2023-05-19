package main

import (
	"log"
	"net"
	"time"

	"github.com/alexflint/go-arg"
)

var args struct {
	SourceAddr string `arg:"-s, --source, required" help:"Source UDP server address (e.g. 0.0.0.0:161)"`
	TargetAddr string `arg:"-t, --target, required" help:"Target UDP server address (e.g. 192.168.9.80:161)"`
}

func init() {
	arg.MustParse(&args)
}

func main() {
	// create port forwarding handler
	fudpAddr, err := net.ResolveUDPAddr("udp", args.TargetAddr)
	if err != nil {
		log.Fatal(err)
	}
	fowardHandler := func(packet []byte) (result []byte, err error) {
		// forward to udp server
		resp, err := forwardServerUdpPacket(fudpAddr, packet)
		if err != nil {
			return nil, err
		}

		return resp, err
	}

	// listen to incoming udp packets, and forward to target udp server
	if err := createUdpServer(args.SourceAddr, fowardHandler); err != nil {
		log.Fatal(err)
	}
}

func forwardServerUdpPacket(addr *net.UDPAddr, packet []byte) (resp []byte, err error) {
	log.Println("Forwarding packet to", addr)

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(60 * time.Second))

	_, err = conn.Write(packet)
	if err != nil {
		return
	}

	// read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	log.Println("Response from forwarding server: ", buf[:n])

	return buf[:n], nil
}

type handler func([]byte) (result []byte, err error)

func createUdpServer(addr string, readHandler ...handler) error {
	// listen to incoming udp packets
	ServerAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	log.Println("Start UDP server Listening on", addr)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		return err
	}
	defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		log.Println("Received from", addr, buf[0:n])
		if err != nil {
			return err
		}

		for _, h := range readHandler {
			// start a goroutine to handle the request
			// no blocking the next request
			go func(h handler, packet []byte) {
				result, err := h(packet)
				if err != nil {
					log.Println("Error: ", err)
				}

				if result != nil {
					ServerConn.WriteToUDP(result, addr)
				}
			}(h, buf[:n])
		}
	}
}
