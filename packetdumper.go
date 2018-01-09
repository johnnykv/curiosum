package main

import (
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func packetDumper(packetMessageChannel chan packetMessage) {

	ethLayer := layers.Ethernet{}
	ipLayer := layers.IPv4{}
	tcpLayer := layers.TCP{}

	handle, err := pcap.OpenLive("en1", 1600, false, 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		parser := gopacket.NewDecodingLayerParser(
			layers.LayerTypeEthernet,
			&ethLayer,
			&ipLayer,
			&tcpLayer,
		)

		foundLayerTypes := []gopacket.LayerType{}

		parser.DecodeLayers(packet.Data(), &foundLayerTypes)

		listenPortes := []uint16{21, 22, 23}

		for _, layerType := range foundLayerTypes {

			if layerType == layers.LayerTypeTCP {
				var getPacket = false
				for _, port := range listenPortes {
					if ((uint16)(tcpLayer.DstPort) == port) || ((uint16)(tcpLayer.SrcPort) == port) {
						getPacket = true

					}
				}
				if getPacket {
					message := packetMessage{}
					message.Timestamp = time.Now()
					message.SYN = tcpLayer.SYN
					message.ACK = tcpLayer.ACK
					message.SrcIP = ipLayer.SrcIP
					message.SrcPort = uint16(tcpLayer.SrcPort)
					message.DstPort = uint16(tcpLayer.DstPort)
					message.DstIP = ipLayer.DstIP
					message.Packet = packet

					packetMessageChannel <- message
				}
			}
		}
	}
}
