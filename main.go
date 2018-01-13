package main

import (
	"flag"
	"os"
	"sync"
)

func main() {

	var captureInterface string
	flag.StringVar(&captureInterface, "i", "", "The interface to listen on")
	flag.Parse()

	if captureInterface == "" {
		flag.Usage()
		os.Exit(2)
	}

	packetMessageChannel := make(chan packetMessage)
	sessionChannel := make(chan sessionMessage)
	pcapWriterChannel := make(chan sessionEntry)

	var wg sync.WaitGroup
	wg.Add(1)
	go heraldingPoller(sessionChannel)
	go pcapWriter(pcapWriterChannel)
	go sessionMaster(&wg, packetMessageChannel, sessionChannel, pcapWriterChannel)
	go packetDumper(packetMessageChannel, captureInterface)
	wg.Wait()
}
