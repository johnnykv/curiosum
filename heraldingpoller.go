package main

import (
	"encoding/json"
	"fmt"
	"strings"

	zmq "github.com/pebbe/zmq4"
)

func heraldingPoller(sessionMessages chan sessionMessage) {
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

		message := sessionMessage{}
		json.Unmarshal([]byte(rawMessage), &message)

		fmt.Printf("Received message type: %s, Content: %v \n", messageType, message)
		sessionMessages <- message
	}

	client.Close()
}
