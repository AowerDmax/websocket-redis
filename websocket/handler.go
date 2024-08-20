package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	RedisClient   *redis.Client
	DataQueueKeys []string
	CheckInterval time.Duration
}

type Message struct {
	Time     string      `json:"time"`
	ListName string      `json:"list_name"`
	Text     string      `json:"text"`
	Id       json.Number `json:"id"`
}

func NewWebSocketHandler(redisClient *redis.Client, dataQueueKeys []string, checkInterval time.Duration) *WebSocketHandler {
	return &WebSocketHandler{
		RedisClient:   redisClient,
		DataQueueKeys: dataQueueKeys,
		CheckInterval: checkInterval,
	}
}

func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	ctx := context.Background()
	lastMessages := make(map[string]Message)

	initialMessages := h.getAllMessages(ctx)
	for _, msg := range initialMessages {
		err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println("Failed to send initial message:", err)
			return
		}
	}

	ticker := time.NewTicker(h.CheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		for _, key := range h.DataQueueKeys {
			lastMessage, err := h.RedisClient.LIndex(ctx, key, -1).Result()
			if err == redis.Nil {
				continue
			} else if err != nil {
				log.Printf("Failed to fetch last item from Redis key %s: %v\n", key, err)
				continue
			}

			var msg Message
			err = json.Unmarshal([]byte(lastMessage), &msg)
			if err != nil {
				log.Printf("Failed to parse message: %v\n", err)
				continue
			}

			if msg != lastMessages[key] {
				messageJSON, err := json.Marshal(msg)
				if err != nil {
					log.Printf("Failed to serialize message: %v\n", err)
					continue
				}

				err = conn.WriteMessage(websocket.TextMessage, messageJSON)
				if err != nil {
					log.Println("Failed to send update message:", err)
					return
				}
				lastMessages[key] = msg
			}
		}
	}
}

func (h *WebSocketHandler) getAllMessages(ctx context.Context) []string {
	var allMessages []string
	for _, key := range h.DataQueueKeys {
		messages, err := h.RedisClient.LRange(ctx, key, 0, -1).Result()
		if err == redis.Nil {
			continue
		} else if err != nil {
			log.Printf("Failed to fetch all messages from Redis key %s: %v\n", key, err)
			continue
		}
		allMessages = append(allMessages, messages...)
	}

	sort.Slice(allMessages, func(i, j int) bool {
		var msg1, msg2 Message
		json.Unmarshal([]byte(allMessages[i]), &msg1)
		json.Unmarshal([]byte(allMessages[j]), &msg2)

		id1, _ := msg1.Id.Int64()
		id2, _ := msg2.Id.Int64()

		return id1 < id2
	})

	return allMessages
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
