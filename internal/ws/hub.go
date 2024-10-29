package ws

type Hub struct {
	broadcast chan []byte
	register chan *Client
	unregister chan *Client
}