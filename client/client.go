package main

import (
	"flag"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type message struct {
	SenderID    string `json:"senderId"`
	RecipientID string `json:"recepientId"`
	Body        string `json:"body"`
	TimeStamp   string `json:"timeStamp"`
}

var addr = flag.String("addr", "localhost:3000", "http service address")
var senderID = flag.String("from", "foo", "sender id")
var recepientID = flag.String("to", "bar", "recepient id")
var messageBody = flag.String("body", "test data", "message body")
var timeInterval = flag.Duration("interval", time.Second*3, "send message time interval in seconds")

func main() {
	flag.Parse()

	var wsURL = url.URL{Scheme: "ws", Host: *addr, Path: "/websockets"}

	log.Printf("connecting to %s", wsURL.String())

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)

	if err != nil {
		log.Fatal("unable to connect:", err)
	}

	if resp.StatusCode != 101 {
		log.Println("server responded with :", resp.StatusCode)
	}

	defer conn.Close()

	go receiveMessages(conn)

	sendMessages(conn)
}

func sendMessages(conn *websocket.Conn) {
	ticker := time.NewTicker(*timeInterval)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			msg := message{*senderID, *recepientID, *messageBody, t.String()}
			err := conn.WriteJSON(msg)

			if err != nil {
				log.Println("Unable to send message:", err)
			}
			log.Println("sent to recepient:", *recepientID)
		}
	}
}

func receiveMessages(conn *websocket.Conn) {
	defer conn.Close()

	for {
		var msg message
		err := conn.ReadJSON(&msg)

		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recieved message from %[1]s. Message: %[2]s", msg.SenderID, msg)
	}
}
