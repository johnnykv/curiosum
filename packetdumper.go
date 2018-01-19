package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func packetDumperWorker(packetMessageChannel chan packetMessage,
	killChannel chan string, captureInterface string, listenPorts []uint16) {

	ethLayer := layers.Ethernet{}
	ipLayer := layers.IPv4{}
	tcpLayer := layers.TCP{}

	handle, err := pcap.OpenLive(captureInterface, 1600, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}

	defer handle.Close()

	var bpfFilter string
	for _, port := range listenPorts {
		bpfFilter = bpfFilter + fmt.Sprintf(" port %d or", port)
	}
	bpfFilter = strings.TrimRight(bpfFilter, " or")
	err = handle.SetBPFFilter(bpfFilter)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Listen interface %s, bpf filter: %v\n", captureInterface, bpfFilter)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	packets := packetSource.Packets()

outer:
	for {
		select {
		case packet := <-packets:
			{
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
		case killMessage := <-killChannel:
			{
				fmt.Println("Killed: " + killMessage)
				break outer
			}
		}
	}

	fmt.Println("Closing packet dumper.")
}

func packetDumper(packetMessageChannel chan packetMessage, captureInterface string, listenPortChannel chan []uint16) {

	var killChannel chan string
	var currentListenPorts []uint16

	for {
		listenPorts := <-listenPortChannel
		sort.Slice(currentListenPorts, func(i, j int) bool { return i < j })
		sort.Slice(listenPorts, func(i, j int) bool { return i < j })

		if reflect.DeepEqual(currentListenPorts, listenPorts) == false {
			currentListenPorts = listenPorts
			// will be nil on first run
			if killChannel != nil {
				killChannel <- "kill"
			}
			killChannel = make(chan string)
			go packetDumperWorker(packetMessageChannel, killChannel, captureInterface, listenPorts)

		}
	}
}
