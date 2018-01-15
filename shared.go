package main

import (
	"net"
	"strconv"
	"time"

	"github.com/google/gopacket"
)

func toHashKey(listenPort uint16, remotePort uint16, ipAddress net.IP) string {
	return strconv.Itoa(int(listenPort)) + "_" + strconv.Itoa(int(remotePort)) + "_" + ipAddress.String()
}

type sessionEntry struct {
	Timestamp time.Time
	Filename  string
	SessionID string
	Packets   []gopacket.Packet
	Heralding bool
}

type sessionMessage struct {
	SessionID  string `json:"sessionID"`
	DstPort    uint16 `json:"DstPort"`
	SrcIP      string `json:"SrcIp"`
	SrcPort    uint16 `json:"SrcPort"`
	SessionEnd bool   `json:"SessionEnded"`
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
