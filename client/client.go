package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"net/url"
	"time"

	"ciphertalk/common/constants"
	"ciphertalk/common/models"

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

	// sendMessages(conn)

	var authToken = login(*addr, *recepientID)

	log.Println("Received auth token:", authToken)
}

func login(host string, user string) string {
	var httpURL = url.URL{Scheme: "http", Host: host, Path: "/login"}
	var loginReq = models.LoginRequest{UserName: user}
	var bodyStr, err = json.Marshal(loginReq)

	if err != nil {
		log.Fatal("unable to convert JSON object to payload")
	}
	var payload = []byte(bodyStr)
	req, err := http.NewRequest(constants.HTTPPost, httpURL.String(), bytes.NewBuffer(payload))

	req.Header.Set(constants.HTTPContentType, constants.HTTPApplicationJSON)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("error happened while sending request:", err)
	}
	defer resp.Body.Close()

	var loginRes = models.LoginResponse{}
	err = json.NewDecoder(resp.Body).Decode(&loginRes)

	if err != nil {
		log.Fatal("unable to parse response from the server")
	}
	return loginRes.AuthToken
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
