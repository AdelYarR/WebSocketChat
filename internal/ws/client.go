package ws

import "github.com/gorilla/websocket"

type Client struct {
	hub  *Hub
	conn *websocket.Conn
}
