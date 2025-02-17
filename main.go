package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"websocket-redis/KnowledgeDatabase"
	"websocket-redis/config"
	"websocket-redis/websocket"

	"github.com/go-redis/redis/v8"
)

func main() {
	cfg := config.LoadConfig()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(20 * time.Second)
		KnowledgeDatabase.ProcessExcel()
	}()

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
	})

	http.HandleFunc("/ws-config.js", func(w http.ResponseWriter, r *http.Request) {
		js := fmt.Sprintf(`var socket = new WebSocket("ws://%s:%d/ws");`, cfg.WsHost, cfg.WsPort)
		w.Header().Set("Content-Type", "application/javascript")
		w.Write([]byte(js))
	})

	wsHandler := websocket.NewWebSocketHandler(rdb, cfg.DataQueueKeys, time.Duration(cfg.IntervalTime)*time.Millisecond)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/ws", wsHandler.ServeWS)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	serverAddr := fmt.Sprintf("%s:%d", cfg.GoAppHost, cfg.GoAppPort)
	fmt.Printf("server start! http://%s \n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
