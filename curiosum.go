package main

import (
	"fmt"
	"sync"
)

func main() {

	fmt.Printf("Running!\n")

	packetMessageChannel := make(chan packetMessage)
	sessionEndChannel := make(chan sessionEndMessage)
	pcapWriterChannel := make(chan sessionEntry)

	var wg sync.WaitGroup
	wg.Add(1)
	go heraldingPoller(sessionEndChannel)
	go pcapWriter(pcapWriterChannel)
	go sessionMaster(&wg, packetMessageChannel, sessionEndChannel, pcapWriterChannel)
	go packetDumper(packetMessageChannel)
	wg.Wait()

}
