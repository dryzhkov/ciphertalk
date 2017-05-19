package models

// Message sent from client to server and transmitted to final recepient
type Message struct {
	SenderID    string `json:"senderId"`
	RecipientID string `json:"recepientId"`
	Body        string `json:"body"`
	TimeStamp   string `json:"timeStamp"`
}

// LoginRequest is sent from client with loging request
type LoginRequest struct {
	UserName string `json:"userName"`
}

// LoginResponse is sent from server in event of successful login
type LoginResponse struct {
	AuthToken string `json:"authToken"`
}
