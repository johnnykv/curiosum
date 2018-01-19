package main

import (
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"sync"
)

func main() {

	var captureInterface string
	var heraldingIPString string
	var heraldingPort uint
	flag.StringVar(&captureInterface, "i", "", "The interface to listen on")
	flag.StringVar(&heraldingIPString, "d", "", "IP address of Heralding, this is used for improved packet filtering")
	flag.UintVar(&heraldingPort, "p", 23400, "The interface to listen on")
	flag.Parse()

	if captureInterface == "" || heraldingPort > math.MaxUint16 {
		flag.Usage()
		os.Exit(2)
	}

	var heraldingIP net.IP
	if heraldingIPString != "" {
		heraldingIP = net.ParseIP(heraldingIPString)
		if heraldingIP == nil {
			fmt.Printf("Could not parse %s as a IP address\n", heraldingIPString)
			os.Exit(2)
		}
	}

	packetMessageChannel := make(chan packetMessage)
	sessionChannel := make(chan sessionMessage)
	pcapWriterChannel := make(chan sessionEntry)
	listenPortChannel := make(chan []uint16)

	var wg sync.WaitGroup
	wg.Add(1)
	go heraldingPoller(sessionChannel, listenPortChannel, uint16(heraldingPort))
	go pcapWriter(pcapWriterChannel)
	go sessionMaster(&wg, packetMessageChannel, sessionChannel, pcapWriterChannel,
		heraldingIP)
	go packetDumper(packetMessageChannel, captureInterface, listenPortChannel)
	wg.Wait()
}
