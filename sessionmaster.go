package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func sessionMaster(wg *sync.WaitGroup, packetMessages chan packetMessage, sessionMessages chan sessionMessage, pcapWriter chan sessionEntry) {
	defer wg.Done()
	var sessions map[string]*sessionEntry
	sessions = make(map[string]*sessionEntry)

	for {
		select {
		case message := <-packetMessages:
			{
				key := toHashKey(message.DstPort, message.SrcPort, message.SrcIP)

				if (message.SYN == true) && (message.ACK == false) {
					entry := sessions[key]
					if entry == nil {
						sessions[key] = &sessionEntry{}
						sessions[key].Timestamp = time.Now()
					} else {
						fmt.Printf("Warning: Entry already exist!\n")
					}
				}

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
		case message := <-sessionMessages:
			{
				key := toHashKey(message.DstPort, message.SrcPort, net.ParseIP(message.SrcIP))
				// this is a session start message
				if message.SessionEnd == false {
					// TODO: check and reset timer
				} else { // this is a session end message
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
}
