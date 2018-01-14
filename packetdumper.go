package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func packetDumper(packetMessageChannel chan packetMessage, captureInterface string, listenPortChannel chan []uint16) {

	ethLayer := layers.Ethernet{}
	ipLayer := layers.IPv4{}
	tcpLayer := layers.TCP{}

	handle, err := pcap.OpenLive(captureInterface, 1600, true, pcap.BlockForever)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// TODO: Receive these from Heralding
	//var listenPorts []uint16
	listenPorts := <-listenPortChannel
	fmt.Printf("Listen interface %s, ports: %v\n", captureInterface, listenPorts)

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
		for _, layerType := range foundLayerTypes {
			if layerType == layers.LayerTypeTCP {
				var getPacket = false
				for _, port := range listenPorts {
					if ((uint16)(tcpLayer.DstPort) == 23) || ((uint16)(tcpLayer.SrcPort) == port) {
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
