package models

// Message sent from client to server and transmitted to final recepient
type Message struct {
	SenderID    string   `json:"senderId"`
	RecipientID string   `json:"recepientId"`
	Body        []byte   `json:"body"`
	TimeStamp   string   `json:"timeStamp"`
	MsgNonce    [24]byte `json:"msgNonce"`
}

// LoginRequest is sent from client with loging request
type LoginRequest struct {
	UserName  string   `json:"userName"`
	PublicKey [32]byte `json:"publicKey"`
}

// LoginResponse is sent from server in event of successful login
type LoginResponse struct {
	AuthToken string `json:"authToken"`
}

// ChannelRequest is sent from client and contains username of another client for whom the channel is being requested
type ChannelRequest struct {
	UserName string `json:"userName"`
}

// ChannelResponse is sent from server and contains public key for the request client
type ChannelResponse struct {
	PublicKey [32]byte `json:"publicKey"`
}
