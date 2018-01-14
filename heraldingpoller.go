package main

import (
	"encoding/json"
	"fmt"
	"strings"

	zmq "github.com/pebbe/zmq4"
)

func heraldingPoller(sessionMessages chan sessionMessage, listenPortMessages chan []uint16) {
	client, err := zmq.NewSocket(zmq.PULL)
	if err != nil {
		panic(err)
	}
	client.Connect("tcp://localhost:23400")
	fmt.Println("Connected")

	for {

		data, err := client.RecvMessage(0)

		if err != nil {
			fmt.Println("Error receiving message")
			break
		}

		result := strings.SplitAfterN(data[0], " ", 2)
		messageType := strings.TrimSpace(result[0])
		rawMessage := result[1]

		fmt.Printf("Received message type: %s, Raw content: %v\n", messageType, rawMessage)
		if messageType == "session_ended" {
			message := sessionMessage{}
			json.Unmarshal([]byte(rawMessage), &message)
			sessionMessages <- message
		} else if messageType == "listen_ports" {
			var listenPorts []uint16
			fmt.Println("SENDIGN")
			json.Unmarshal([]byte(rawMessage), &listenPorts)
			listenPortMessages <- listenPorts
		} else {
			fmt.Printf("Unknown message received, raw data: %v", data)
		}
	}

	client.Close()
}
