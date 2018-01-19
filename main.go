package main

import (
	"flag"
	"math"
	"os"
	"sync"
)

func main() {

	var captureInterface string
	var heraldingPort uint
	flag.StringVar(&captureInterface, "i", "", "The interface to listen on")
	flag.UintVar(&heraldingPort, "p", 23400, "The interface to listen on")
	flag.Parse()

	if captureInterface == "" || heraldingPort > math.MaxUint16 {
		flag.Usage()
		os.Exit(2)
	}

	packetMessageChannel := make(chan packetMessage)
	sessionChannel := make(chan sessionMessage)
	pcapWriterChannel := make(chan sessionEntry)
	listenPortChannel := make(chan []uint16)

	var wg sync.WaitGroup
	wg.Add(1)
	go heraldingPoller(sessionChannel, listenPortChannel, uint16(heraldingPort))
	go pcapWriter(pcapWriterChannel)
	go sessionMaster(&wg, packetMessageChannel, sessionChannel, pcapWriterChannel)
	go packetDumper(packetMessageChannel, captureInterface, listenPortChannel)
	wg.Wait()
}
