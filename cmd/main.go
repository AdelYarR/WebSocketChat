package main

import (
	"net/http"

	"github.com/AdelYarR/WebSocketChat/internal/ws"
)

func main() {
	hub := ws.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(hub, w, r)
	})

	http.ListenAndServe("localhost:8080", nil)
}
