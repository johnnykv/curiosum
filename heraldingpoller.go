package main

import (
	"encoding/json"
	"fmt"
	"strings"

	zmq "github.com/pebbe/zmq4"
)

func heraldingPoller(sessionMessages chan sessionMessage, listenPortMessages chan []uint16, heraldingPort uint16) {
	client, err := zmq.NewSocket(zmq.PULL)
	if err != nil {
		panic(err)
	}
	socketURL := fmt.Sprintf("tcp://localhost:%d", heraldingPort)
	err = client.Connect(socketURL)

	if err != nil {
		fmt.Printf("Error connecting: %v\n", err)
		return
	}

	fmt.Printf("Connecting to Heralding instance on %s\n", socketURL)

	for {

		data, err := client.RecvMessage(0)

		if err != nil {
			fmt.Printf("Error receiving message: %v\n", err)
			break
		}

		result := strings.SplitAfterN(data[0], " ", 2)
		messageType := strings.TrimSpace(result[0])
		rawMessage := result[1]

		//fmt.Printf("Received message type: %s, Raw content: %v\n", messageType, rawMessage)
		if messageType == "session_ended" {
			message := sessionMessage{}
			json.Unmarshal([]byte(rawMessage), &message)
			sessionMessages <- message
		} else if messageType == "listen_ports" {
			var listenPorts []uint16
			json.Unmarshal([]byte(rawMessage), &listenPorts)
			listenPortMessages <- listenPorts
		} else {
			fmt.Printf("Unknown message received, raw data: %v", data)
		}
	}

	client.Close()
}
