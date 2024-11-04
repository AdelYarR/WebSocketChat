package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/AdelYarR/WebSocketChat/internal/models"
	"github.com/AdelYarR/WebSocketChat/utils"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	id   string
}

func (c *Client) writeMsg() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) readMsg(Rclient *redis.Client) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		msgInfo := models.Message{
			Sender: c.id,
			Date: time.Now().Format("2006-01-02 15:04:05"),
			Text: string(message),
		}

		encodedMsg, err := json.Marshal(msgInfo)
		if err != nil {
			return
		}

		Rclient.RPush(context.Background(), c.hub.id, encodedMsg)

		c.hub.broadcast <- []byte(encodedMsg)
	}
}

func ServeWS(Rclient *redis.Client, hubMap map[string]*Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	u := r.URL
	var hub *Hub
	params := u.Query()
	joinRoom := params.Get("joinRoom")

	if joinRoom == "" {
		log.Printf("Requested URL: %s", r.URL.String())
		log.Printf("Query parameters: %v", r.URL.RawQuery)
		return
	}

	if _, ok := hubMap[joinRoom]; !ok {
		hub = NewHub(joinRoom)
		go hub.Run()
		hubMap[joinRoom] = hub
	} else {
		hub = hubMap[joinRoom]
	}

	client_id, err := utils.GenerateHashId()
	if err != nil {
		log.Printf("failed to generate client id " + err.Error())
		return
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		id:   client_id,
	}

	client.hub.register <- client

	go client.writeMsg()
	go client.readMsg(Rclient)
	
	err = loadMessages(Rclient, client.hub)
	if err != nil {
		log.Printf("failed to load messages: " + err.Error())
		return
	}
}

func loadMessages(Rclient *redis.Client, hub *Hub) error {
	messages, err := Rclient.LRange(context.Background(), hub.id, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, msg := range messages {
		hub.broadcast <- []byte(msg)
	}

	return nil
}
