package main

import (
	"fmt"
	"os"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

func pcapWriter(pcapWriter chan sessionEntry) {
	for {
		sessionEntry := <-pcapWriter
		fileName := sessionEntry.SessionID + ".pcap"
		fmt.Printf("Writing session to %s\n", fileName)
		file, _ := os.Create(fileName)
		defer file.Close()
		writer := pcapgo.NewWriter(file)
		writer.WriteFileHeader(1600, layers.LinkTypeEthernet)
		for _, packet := range sessionEntry.Packets {
			writer.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
		}
	}
}
