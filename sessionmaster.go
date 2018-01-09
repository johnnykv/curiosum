package main

import (
	"fmt"
	"net"
	"sync"
)

func sessionMaster(wg *sync.WaitGroup, packetMessages chan packetMessage, sessionEndMessages chan sessionEndMessage, pcapWriter chan sessionEntry) {
	defer wg.Done()
	var sessions map[string]*sessionEntry
	sessions = make(map[string]*sessionEntry)

	for {
		select {
		case message := <-packetMessages:
			{
				key := ""
				if (message.SYN == true) && (message.ACK == false) {
					key = toHashKey(message.DstPort, message.SrcPort, message.SrcIP)
					entry := sessions[key]
					// TODO: If not nil something is wrong or corner case...
					if entry == nil {
						sessions[key] = &sessionEntry{}
					} else {
						fmt.Printf("Warning: Entry already exist!\n")
					}
				}

				key = toHashKey(message.DstPort, message.SrcPort, message.SrcIP)
				entry := sessions[key]
				if entry != nil {
					entry.Packets = append(entry.Packets, message.Packet)
				} else {
					keyTwo := toHashKey(message.SrcPort, message.DstPort, message.DstIP)
					entry = sessions[keyTwo]
					if entry != nil {
						entry.Packets = append(entry.Packets, message.Packet)
					}
				}
			}

		case message := <-sessionEndMessages:
			{
				key := toHashKey(message.DstPort, message.SrcPort, net.ParseIP(message.SrcIP))
				entry := sessions[key]
				if entry != nil {
					entry.SessionID = message.SessionID
					pcapWriter <- *entry
					delete(sessions, key)
				} else {
					fmt.Printf("Not found: %s!\n", key)
				}
			}
		}
	}
}
