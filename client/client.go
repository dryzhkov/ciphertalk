package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/crypto/nacl/box"

	"ciphertalk/common/constants"
	"ciphertalk/common/models"

	"github.com/gorilla/websocket"
)

type keys struct {
	publicKey  [32]byte
	privateKey [32]byte
}

var addr = flag.String("addr", "localhost:3000", "http service address")
var senderID = flag.String("from", "foo", "sender id")
var recepientID = flag.String("to", "bar", "recepient id")
var messageBody = flag.String("body", "test data", "message body")
var timeInterval = flag.Duration("interval", time.Second*3, "send message time interval in seconds")
var listenOnly = flag.Bool("listen-only", false, "client will not send any messages")
var myKeys keys

func main() {
	flag.Parse()
	// generate a new public/private key pair
	myKeys = generateKeys()

	// get auth token
	authToken := login(*addr, *senderID, &myKeys)

	// get recepient's public key (create secure channel)
	recepientPubKey, err := getRecipientKey(*addr, authToken, *recepientID)

	for err != nil {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("User with name [" + *recepientID + "] has not registered yet. Register the user first and press enter to continue...")
		reader.ReadString('\n')
		recepientPubKey, err = getRecipientKey(*addr, authToken, *recepientID)
	}

	log.Println("recepient pub key ", recepientPubKey)

	conn := openWebsocket(&authToken)
	defer conn.Close()

	if !*listenOnly {
		go sendMessages(conn, &recepientPubKey)
	}

	receiveMessages(conn, &recepientPubKey)
}

func openWebsocket(authToken *string) *websocket.Conn {
	wsURL := url.URL{Scheme: "ws", Host: *addr, Path: "/websockets"}
	headers := http.Header{
		constants.HTTPAuthorization: {fmt.Sprintf("Bearer %v", *authToken)},
	}

	log.Printf("connecting to %s", wsURL.String())

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL.String(), headers)

	if err != nil {
		log.Fatal("unable to connect via websocket:", err)
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		log.Println("server responded with :", resp.StatusCode)
	}

	return conn
}

func login(host string, user string, k *keys) string {
	var httpURL = url.URL{Scheme: "http", Host: host, Path: "/login"}
	var loginReq = models.LoginRequest{UserName: user, PublicKey: k.publicKey}
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

func getRecipientKey(host string, authToken string, recepientID string) ([32]byte, error) {
	var httpURL = url.URL{Scheme: "http", Host: host, Path: "/secure"}
	var chReq = models.ChannelRequest{UserName: recepientID}
	var bodyStr, err = json.Marshal(chReq)

	if err != nil {
		log.Fatal("unable to convert JSON object to payload")
	}
	var payload = []byte(bodyStr)
	req, err := http.NewRequest(constants.HTTPPost, httpURL.String(), bytes.NewBuffer(payload))

	req.Header.Set(constants.HTTPAuthorization, fmt.Sprintf("Bearer %v", authToken))
	req.Header.Set(constants.HTTPContentType, constants.HTTPApplicationJSON)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("error happened while sending request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return [32]byte{}, errors.New("Client not registered")
	}

	var chRes = models.ChannelResponse{}
	err = json.NewDecoder(resp.Body).Decode(&chRes)

	if err != nil {
		log.Fatal("unable to parse response from the server")
	}
	return chRes.PublicKey, nil
}

func sendMessages(conn *websocket.Conn, recepientKey *[32]byte) {
	ticker := time.NewTicker(*timeInterval)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			msgBytes := []byte(*messageBody)
			encyptedMsg := encrypt(&msgBytes, &myKeys, recepientKey, t.String())
			err := conn.WriteJSON(encyptedMsg)

			if err != nil {
				log.Println("Unable to send message:", err)
			}
			log.Printf("sent to recepient:%[1]v message: %[2]v\n", *recepientID, *messageBody)
		}
	}
}

func receiveMessages(conn *websocket.Conn, recepientKey *[32]byte) {
	defer conn.Close()

	for {
		var msg models.Message
		err := conn.ReadJSON(&msg)

		if err != nil {
			log.Println("read:", err)
			return
		}

		decryptAndPrint(msg, &myKeys, recepientKey)
	}
}

func generateKeys() keys {
	pubKey, priKey, err := box.GenerateKey(rand.Reader)

	if err != nil {
		log.Fatal(err)
	}

	return keys{publicKey: *pubKey, privateKey: *priKey}
}

func encrypt(msgBytes *[]byte, myKeys *keys, recepientKey *[32]byte, timeStamp string) models.Message {
	var out []byte
	var nonce [24]byte
	randomizeNonce(&nonce)
	encryptedBytes := box.Seal(out, *msgBytes, &nonce, recepientKey, &myKeys.privateKey)

	return models.Message{
		SenderID:    *senderID,
		RecipientID: *recepientID,
		Body:        encryptedBytes,
		TimeStamp:   timeStamp,
		MsgNonce:    nonce,
	}
}

func decryptAndPrint(msg models.Message, myKeys *keys, recepientKey *[32]byte) {
	var out []byte
	decryptedBytes, success := box.Open(out, msg.Body, &msg.MsgNonce, recepientKey, &myKeys.privateKey)

	if !success {
		log.Printf("Something went wrong... unable to decrypt message: %[1]s", decryptedBytes)
	} else {
		decryptedMsg := string(decryptedBytes[:len(decryptedBytes)])
		log.Printf("recieved message from %[1]s. Message: %[2]s", msg.SenderID, decryptedMsg)
	}
}

func randomizeNonce(nonce *[24]byte) {
	b := make([]byte, 1)
	for i := 0; i < len(nonce); i++ {
		rand.Read(b)
		nonce[i] = b[0]
	}
}
