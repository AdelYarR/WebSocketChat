package main

import (
	"net/http"

	"github.com/AdelYarR/WebSocketChat/internal/ws"
)

func main() {
	hubMap := make(map[string]*ws.Hub)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(hubMap, w, r)
	})

	http.ListenAndServe("localhost:8080", nil)
}
