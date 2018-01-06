package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	zmq "github.com/pebbe/zmq4"
)

type sessionEntry struct {
	Timestamp time.Time
	Packets   []gopacket.Packet
}

type sessionStartMessage struct {
	SessionID string `json:"sessionID"`
	DstPort   uint16 `json:"DstPort"`
	SrcIP     string `json:"SrcIp"`
	SrcPort   uint16 `json:"SrcPort"`
}

type sessionEndMessage struct {
	SessionID string `json:"sessionID"`
	DstPort   uint16 `json:"DstPort"`
	SrcIP     string `json:"SrcIp"`
	SrcPort   uint16 `json:"SrcPort"`
}

type packetMessage struct {
	Timestamp time.Time
	SYN       bool
	ACK       bool
	SrcIP     net.IP
	SrcPort   uint16
	DstIP     net.IP
	DstPort   uint16
	Packet    gopacket.Packet
}

func heraldingPoller() {
	client, err := zmq.NewSocket(zmq.PULL)
	if err != nil {
		panic(err)
	}
	client.Connect("tcp://localhost:23400")
	fmt.Println("Connected")

	for {

		reply, err := client.RecvMessage(0)

		if err != nil {
			fmt.Println("Error receiving message")
			break
		}

		result := strings.SplitAfterN(reply[0], " ", 2)
		messageType := result[0]
		rawMessage := result[1]

		message := sessionEndMessage{}
		json.Unmarshal([]byte(rawMessage), &message)

		fmt.Printf("Received message type: %s, Content: %v \n", messageType, message)
	}

	client.Close()
}

func toHashKey(listenPort uint16, remotePort uint16, ipAddress net.IP) string {
	return strconv.Itoa(int(listenPort)) + "_" + strconv.Itoa(int(remotePort)) + "_" + ipAddress.String()
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

func sessionMaster(packetMessages chan packetMessage) {
	var sessions map[string]*sessionEntry
	sessions = make(map[string]*sessionEntry)

	// TODO: IF session ended received send data to file writer using
	/*for k, v := range sessions {
		fmt.Printf("Writing %v packets from %s\n", len(v.Packets), k)
		writePcapFile(v.Packets, k+".pcap")
	}*/

	for message := range packetMessages {
		//fmt.Println(message)
		key := ""
		if (message.SYN == true) && (message.ACK == false) {
			key = toHashKey(message.DstPort, message.SrcPort, message.SrcIP)
			fmt.Printf("Adding array for: %s\n", key)
			entry := sessions[key]
			// TODO: If not nil something is wrong or corner case...
			if entry == nil {
				sessions[key] = &sessionEntry{}
			} else {
				fmt.Printf("Warning: Entry already existed!\n")
			}
		}

		key = toHashKey(message.DstPort, message.SrcPort, message.SrcIP)
		entry := sessions[key]
		if entry != nil {
			fmt.Printf("Adding packet to key: %s\n", key)
			entry.Packets = append(entry.Packets, message.Packet)
			fmt.Printf("Len of array: %v\n", len(sessions[key].Packets))
		} else {
			keyTwo := toHashKey(message.SrcPort, message.DstPort, message.DstIP)
			entry = sessions[keyTwo]
			if entry != nil {

				fmt.Printf("Adding packet to key two: %s\n", keyTwo)
				entry.Packets = append(entry.Packets, message.Packet)
				fmt.Printf("Len of array: %v\n", len(sessions[keyTwo].Packets))
			}
		}
		fmt.Printf("Len of hashmap: %v\n", len(sessions))
	}
}

func main() {

	handle, err := pcap.OpenLive("en1", 1600, false, 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()
	fmt.Printf("Running!\n")
	go heraldingPoller()
	ethLayer := layers.Ethernet{}
	ipLayer := layers.IPv4{}
	tcpLayer := layers.TCP{}

	x := 0

	packetMessageChannel := make(chan packetMessage)
	go sessionMaster(packetMessageChannel)

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

		// while developing...
		if x > 20 {
			break
		}

		for _, layerType := range foundLayerTypes {
			if layerType == layers.LayerTypeTCP {

				var getPacket = false
				for _, port := range listenPortes {
					if ((uint16)(tcpLayer.DstPort) == port) || ((uint16)(tcpLayer.SrcPort) == port) {
						getPacket = true
					}
				}
				if getPacket {
					x++
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
