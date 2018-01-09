package main

import (
	"fmt"
	"os"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

func pcapWriter(pcapWriter chan sessionEntry) {
	for {
		//reply, err := sessionEntry.RecvMessage(0)
		sessionEntry := <-pcapWriter
		fileName := sessionEntry.SessionID + ".pcap"
		fmt.Printf("Writing to %s", fileName)
		file, _ := os.Create(fileName)
		writer := pcapgo.NewWriter(file)
		writer.WriteFileHeader(1600, layers.LinkTypeEthernet)
		defer file.Close()
		for _, packet := range sessionEntry.Packets {
			writer.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
		}
	}
}
