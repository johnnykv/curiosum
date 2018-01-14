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

	killChannel := make(chan string)

	// discard session if no notification from Heralding within timeout.
	// if Heralding has notified us of the session, the timeout will be doubled
	var sessionTimeoutSeconds float64
	sessionTimeoutSeconds = 5

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
						sessions[key].Heralding = false
						go entryKiller(sessions[key], key, killChannel, sessionTimeoutSeconds)
					} else {
						fmt.Printf("Warning: Entry already exist: %s\n", key)
					}
				}

				entry := sessions[key]
				if entry != nil {
					entry.Packets = append(entry.Packets, message.Packet)
					entry.Timestamp = time.Now()
				} else {
					keyTwo := toHashKey(message.SrcPort, message.DstPort, message.DstIP)
					entry = sessions[keyTwo]
					if entry != nil {
						entry.Packets = append(entry.Packets, message.Packet)
						entry.Timestamp = time.Now()
					}
				}
			}
		case message := <-sessionMessages:
			{
				key := toHashKey(message.DstPort, message.SrcPort, net.ParseIP(message.SrcIP))
				// this is a session start message
				if message.SessionEnd == false {
					entry := sessions[key]
					if entry != nil {
						entry.Heralding = true
					}
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
		case key := <-killChannel:
			{
				fmt.Printf("Kill signal received for %v\n", key)
				fmt.Println(len(sessions))
				delete(sessions, key)
				fmt.Println(len(sessions))

			}
		}
	}
}

func entryKiller(entry *sessionEntry, key string, killChannel chan string, timeoutSeconds float64) {
	var timeout float64
	timeout = timeoutSeconds
	if entry.Heralding == true {
		timeout = timeout * 2
	}
	for {
		time.Sleep(1 * time.Second)
		if (time.Now().Sub(entry.Timestamp)).Seconds() > timeout {
			killChannel <- key
			break
		}
	}
}
