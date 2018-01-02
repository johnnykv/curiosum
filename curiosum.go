package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
)

func toHashKey(listenPort layers.TCPPort, remotePort layers.TCPPort, ipAddress net.IP) string {
	return listenPort.String() + "_" + remotePort.String() + "_" + ipAddress.String()
}

func writePcapFile(entries []gopacket.Packet, fileName string) {
	file, _ := os.Create(fileName)
	writer := pcapgo.NewWriter(file)
	writer.WriteFileHeader(1600, layers.LinkTypeEthernet)
	defer file.Close()
	for _, packet := range entries {
		writer.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
	}
}

func main() {

	handle, err := pcap.OpenLive("en1", 1600, false, 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()
	fmt.Printf("Running!\n")

	ethLayer := layers.Ethernet{}
	ipLayer := layers.IPv4{}
	tcpLayer := layers.TCP{}

	var sessions map[string][]gopacket.Packet
	sessions = make(map[string][]gopacket.Packet)
	x := 0
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		parser := gopacket.NewDecodingLayerParser(
			layers.LayerTypeEthernet,
			&ethLayer,
			&ipLayer,
			&tcpLayer,
		)

		foundLayerTypes := []gopacket.LayerType{}

		err := parser.DecodeLayers(packet.Data(), &foundLayerTypes)
		if err != nil {
			fmt.Printf("Trouble decoding layer!\n")
		}

		// while developing...
		if x > 20 {
			break
		}

		for _, layerType := range foundLayerTypes {
			if layerType == layers.LayerTypeTCP {

				// TODO: if port in range we are interested in...
				if (tcpLayer.DstPort == 23) || (tcpLayer.SrcPort == 23) {
					x++

					// from remote to us
					key := toHashKey(tcpLayer.DstPort, tcpLayer.SrcPort, ipLayer.SrcIP)

					if (tcpLayer.SYN == true) && (tcpLayer.ACK == false) {
						fmt.Printf("Adding array for: %s\n", key)
						entry := sessions[key]
						// TODO: If not nil something is wrong or corner case...
						if entry == nil {
							sessions[key] = []gopacket.Packet{}
						}
					}

					entry := sessions[key]
					if entry != nil {
						fmt.Printf("Adding packet to key: %s\n", key)
						sessions[key] = append(entry, packet)
						fmt.Printf("Len of array: %v\n", len(sessions[key]))
					} else {
						keyTwo := toHashKey(tcpLayer.SrcPort, tcpLayer.DstPort, ipLayer.DstIP)
						entry = sessions[keyTwo]
						if entry != nil {

							fmt.Printf("Adding packet to key two: %s\n", keyTwo)
							sessions[keyTwo] = append(entry, packet)
							fmt.Printf("Len of array: %v\n", len(sessions[keyTwo]))
						}
					}
					fmt.Printf("Len of hashmap: %v\n", len(sessions))

				}
			}
		}
	}

	for k, v := range sessions {
		fmt.Printf("Writing %v packets from %s\n", len(v), k)
		writePcapFile(v, k+".pcap")
	}
}
