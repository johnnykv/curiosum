package main

import (
	"fmt"
	"sync"
)

func main() {

	fmt.Printf("Running!\n")

	packetMessageChannel := make(chan packetMessage)
	sessionChannel := make(chan sessionMessage)
	pcapWriterChannel := make(chan sessionEntry)

	var wg sync.WaitGroup
	wg.Add(1)
	go heraldingPoller(sessionChannel)
	go pcapWriter(pcapWriterChannel)
	go sessionMaster(&wg, packetMessageChannel, sessionChannel, pcapWriterChannel)
	go packetDumper(packetMessageChannel)
	wg.Wait()
}
