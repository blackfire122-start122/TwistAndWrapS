package internal

import (
	. "TwistAndWrapS/pkg"
	. "TwistAndWrapS/pkg/logging"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type ClientBar struct {
	Conn   *websocket.Conn
	Bar    Bar
	RoomId string
}

type Message struct {
	Type   string          `json:"Type"`
	RoomId string          `json:"RoomId,omitempty"`
	Data   json.RawMessage `json:"Data"`
}

type CreateOrderMessage struct {
	Client          *ClientBar `json:"Client,omitempty"`
	ProductsCreated []uint64   `json:"ProductsCreated"`
	Id              uint64     `json:"Id"`
}

type OrderGiveMessage struct {
	Id uint64 `json:"Id"`
}

var Clients = make(map[*ClientBar]bool)
var Broadcast = make(chan *Message)
var BroadcastCreateOrder = make(chan *CreateOrderMessage)

func deleteClient(client *ClientBar) {
	delete(Clients, client)
	err := ClientRedis.Del(context.Background(), strconv.FormatUint(client.Bar.Id, 10))
	if err != nil {
		ErrorLogger.Println("Error deleting key from Redis:", err)
	}
}

func sendInRedisWebsocketChannel(msg interface{}) {
	result := ClientRedis.Publish(Ctx, WebsocketChannel, msg)
	if result.Err() != nil {
		ErrorLogger.Println("Error publishing message:", result.Err())
		fmt.Println("Publish result:", result)
		if errors.Is(result.Err(), redis.ErrClosed) || strings.Contains(result.Err().Error(), "actively refused") {
			_ = ReconnectToRedis()
		} else {
			ErrorLogger.Println("Error publishing message:", result.Err())
		}
	}

}

func receiver(client *ClientBar) {
	defer func() {
		if err := client.Conn.Close(); err != nil {
			ErrorLogger.Println(err.Error())
		}
	}()

	for {
		_, p, err := client.Conn.ReadMessage()
		if err != nil {
			deleteClient(client)
			ErrorLogger.Println("Error read message: " + err.Error())
			break
		}

		var m Message
		err = json.Unmarshal(p, &m)
		if err != nil {
			ErrorLogger.Println(err.Error())
		}

		m.RoomId = client.RoomId

		msg, err := json.Marshal(&m)
		if err != nil {
			ErrorLogger.Println(err.Error())
			continue
		}

		sendInRedisWebsocketChannel(msg)
	}
}

func Broadcaster() {
	for {
		msg := <-Broadcast
		for client := range Clients {
			if client.RoomId == msg.RoomId {
				err := client.Conn.WriteJSON(Message{Type: msg.Type, Data: msg.Data})

				if err != nil {
					ErrorLogger.Println("Error write message: " + err.Error())
					deleteClient(client)
					err := client.Conn.Close()
					if err != nil {
						ErrorLogger.Println("Error close message: " + err.Error())
					}
					return
				}
				break
			}
		}
	}
}

func WsChat(c *gin.Context) {
	res, bar := CheckSessionBar(c.Request)
	if !res {
		return
	}

	roomId := c.Request.URL.Query().Get("roomId")

	if roomId == "" {
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		ErrorLogger.Println(err.Error())
		return
	}

	client := &ClientBar{Conn: conn, RoomId: roomId, Bar: bar}
	Clients[client] = true

	jsonData, err := json.Marshal(BarSessionInRedis{
		RemoteAddr: conn.RemoteAddr().String(),
		RoomId:     roomId,
	})

	if err != nil {
		ErrorLogger.Println(err.Error())
	}

	err = ClientRedis.Set(c.Request.Context(), strconv.FormatUint(bar.Id, 10), jsonData, 0).Err()
	if err != nil {
		log.Println("Error storing session in Redis:", err)
		return
	}

	go receiver(client)
	go pingPong(client)
}

func pingPong(client *ClientBar) {
	ticker := time.NewTicker(30 * time.Second)

	defer func() {
		ticker.Stop()
		err := client.Conn.Close()
		if err != nil {
			ErrorLogger.Println("Error close", err)
		}
	}()

	for range ticker.C {
		if err := client.Conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second)); err != nil {
			ErrorLogger.Println("Error sending ping message: " + err.Error())
			break
		}
	}
}

func RedisReceiver() {
	PubSub := ClientRedis.Subscribe(Ctx, WebsocketChannel)
	defer func(PubSub *redis.PubSub) {
		err := PubSub.Close()
		if err != nil {
			ErrorLogger.Println(err.Error())
		}
	}(PubSub)

	for {
		msg, err := PubSub.ReceiveMessage(Ctx)
		if err != nil {
			ErrorLogger.Println(err.Error())
			return
		}

		var m Message
		err = json.Unmarshal([]byte(msg.Payload), &m)
		if err != nil {
			ErrorLogger.Println(err.Error())
		}

		var client *ClientBar

		for cl := range Clients {
			if cl.RoomId == m.RoomId {
				client = cl
				break
			}
		}

		if client == nil {
			continue
		}

		if m.Type == "OrderCreated" {
			var msg CreateOrderMessage
			err := json.Unmarshal(m.Data, &msg)
			if err != nil {
				ErrorLogger.Println("Error unmarshal create order msg: ", err)
				continue
			}
			msg.Client = client
			BroadcastCreateOrder <- &msg
		} else if m.Type == "OrderGive" {
			var msg OrderGiveMessage
			err := json.Unmarshal(m.Data, &msg)
			if err != nil {
				ErrorLogger.Println("Error unmarshal create order msg: ", err)
				continue
			}

			var order Order
			if err := DB.Preload("OrderProducts").First(&order, "order_id=?", msg.Id).Error; err != nil {
				ErrorLogger.Println("Error not found order ", err)
				continue
			}
			for _, orderProduct := range order.OrderProducts {
				if err := DB.Unscoped().Delete(&orderProduct).Error; err != nil {
					ErrorLogger.Println("Error deleting order product: ", err)
					continue
				}
			}
			if err := DB.Unscoped().Delete(&order).Error; err != nil {
				ErrorLogger.Println("Error delete order ", err)
			}
		}
		if m.Type == "createOrder" {
			Broadcast <- &m
		}
	}
}
