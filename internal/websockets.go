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
	Type   string     `json:"Type"`
	Msg    string     `json:"Msg"`
	Client *ClientBar `json:"Client,omitempty"`
}

type MessageChat struct {
	Type string
	Msg  string
}

type RespMessage struct {
	Type            string
	Id              uint64
	ProductsCreated []uint64
	Msg             string
	Client          *ClientBar
}

var Clients = make(map[*ClientBar]bool)
var Broadcast = make(chan *Message)
var BroadcastReceiver = make(chan *RespMessage)

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

		var m RespMessage

		err = json.Unmarshal(p, &m)
		if err != nil {
			ErrorLogger.Println(err.Error())
			break
		}

		m.Client = client

		if m.Type == "OrderCreated" {
			BroadcastReceiver <- &m
		}
	}
}

func Broadcaster() {
	for {
		message := <-Broadcast
		for client := range Clients {
			if client.RoomId == message.Client.RoomId {
				err := client.Conn.WriteJSON(MessageChat{Type: message.Type, Msg: message.Msg})

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
