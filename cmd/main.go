package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AdelYarR/WebSocketChat/config"
	"github.com/AdelYarR/WebSocketChat/internal/ws"
	"github.com/go-redis/redis/v8"
)

func main() {
	cfg := config.MustLoad()
	hubMap := make(map[string]*ws.Hub)

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: "",
		DB:       0,
	})

	ping, err := client.Ping(context.Background()).Result()
	if err != nil {
		fmt.Println("error connecting to redis: " + err.Error())
		return
	}

	fmt.Println(ping)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(client, hubMap, w, r)
	})

	http.ListenAndServe(cfg.BindAddr, nil)
}
