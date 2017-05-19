package controller

import (
	"ciphertalk/common/models"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type client struct {
	id     string
	socket *websocket.Conn
}

// APIController represents API controller
type APIController struct {
	clients  []client
	mutex    sync.Mutex
	upgrader websocket.Upgrader
	channel  chan models.Message
}

// NewAPIController creates new instance of APIController
func NewAPIController() *APIController {
	ctrl := new(APIController)

	ctrl.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ctrl.channel = make(chan models.Message)

	go ctrl.processMessages()

	return ctrl
}

// HandleWebsockets saves incoming connections, reads messages and notifies message handler via a channel
func (ctrl *APIController) HandleWebsockets(w http.ResponseWriter, r *http.Request) {
	socket, err := ctrl.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	cl := client{socket: socket, id: ""}

	for {
		var msg models.Message

		err := socket.ReadJSON(&msg)

		if err != nil {
			log.Printf("Unexpected error parsing message: %v", err)
			ctrl.removeClient(cl)
			break
		}

		if !ctrl.isValid(&msg) {
			ctrl.removeClient(cl)
			break
		}

		if cl.id == "" {
			// first message from this client
			cl.id = msg.SenderID
			ctrl.addClient(cl)
		}

		log.Printf("Recieved message from: %[1]v\n", msg.SenderID)
		ctrl.channel <- msg
	}
}

func (ctrl *APIController) isValid(msg *models.Message) bool {
	if msg.RecipientID == "" {
		log.Printf("Invalid recepient")
		return false
	}

	if msg.SenderID == "" {
		log.Printf("Invalid sender")
		return false
	}

	if msg.Body == "" {
		log.Printf("Invalid message body")
		return false
	}

	return true
}

// Sends incoming message to correct client
// If recepient is offline, removes it from the list of clients
func (ctrl *APIController) processMessages() {
	for {
		msg := <-ctrl.channel

		for _, cl := range ctrl.clients {

			if cl.id == msg.RecipientID {
				log.Printf("Sending message to: %[1]v\n", msg.RecipientID)
				err := cl.socket.WriteJSON(msg)

				if err != nil {
					ctrl.removeClient(cl)
				}
			}
		}
	}
}

func (ctrl *APIController) addClient(c client) {
	ctrl.mutex.Lock()
	ctrl.clients = append(ctrl.clients, c)
	ctrl.mutex.Unlock()
}

func (ctrl *APIController) removeClient(c client) {
	ctrl.mutex.Lock()
	c.socket.Close()

	for i := range ctrl.clients {
		if ctrl.clients[i].id == c.id {
			ctrl.clients = append(ctrl.clients[:i], ctrl.clients[i+1:]...)
		}
	}

	ctrl.mutex.Unlock()
}
