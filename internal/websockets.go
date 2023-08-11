package internal

import (
	. "TwistAndWrapS/pkg"
	. "TwistAndWrapS/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
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
	Type string          `json:"Type"`
	Data json.RawMessage `json:"Data"`
}

type ClientMessage struct {
	Client  *ClientBar
	Message *Message
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
var Broadcast = make(chan *ClientMessage)
var BroadcastCreateOrder = make(chan *CreateOrderMessage)

func receiver(client *ClientBar) {
	defer func() {
		if err := client.Conn.Close(); err != nil {
			ErrorLogger.Println(err.Error())
		}
	}()

	for {
		_, p, err := client.Conn.ReadMessage()
		if err != nil {
			delete(Clients, client)
			ErrorLogger.Println("Error read message: " + err.Error())
			break
		}

		var m Message

		err = json.Unmarshal(p, &m)
		if err != nil {
			ErrorLogger.Println(err.Error())
			break
		}

		if m.Type == "OrderCreated" {
			var msg CreateOrderMessage
			err := json.Unmarshal(m.Data, &msg)
			if err != nil {
				ErrorLogger.Println("Error unmarshal create order msg: ", err)
				return
			}
			msg.Client = client
			BroadcastCreateOrder <- &msg
		} else if m.Type == "OrderGive" {
			var msg OrderGiveMessage
			err := json.Unmarshal(m.Data, &msg)
			if err != nil {
				ErrorLogger.Println("Error unmarshal create order msg: ", err)
				return
			}

			var order Order
			if err := DB.Preload("OrderProducts").First(&order, "order_id=?", msg.Id).Error; err != nil {
				ErrorLogger.Println("Error not found order ", err)
			}
			for _, orderProduct := range order.OrderProducts {
				if err := DB.Unscoped().Delete(&orderProduct).Error; err != nil {
					ErrorLogger.Println("Error deleting order product: ", err)
				}
			}
			if err := DB.Unscoped().Delete(&order).Error; err != nil {
				ErrorLogger.Println("Error delete order ", err)
			}
		}
	}
}

func Broadcaster() {
	for {
		msgCl := <-Broadcast
		for client := range Clients {
			if client.RoomId == msgCl.Client.RoomId {
				err := client.Conn.WriteJSON(Message{Type: msgCl.Message.Type, Data: msgCl.Message.Data})

				if err != nil {

					fmt.Println("err write message ", err)

					ErrorLogger.Println("Error write message: " + err.Error())
					delete(Clients, client)
					err := client.Conn.Close()
					if err != nil {
						ErrorLogger.Println("Error close message: " + err.Error())
					}
					return
				}
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

	client := ClientBar{Conn: conn, RoomId: roomId, Bar: bar}
	Clients[&client] = true

	go receiver(&client)
	go pingPong(&client)
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
