package internal

import (
	. "TwistAndWrapS/pkg"
	. "TwistAndWrapS/pkg/logging"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
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

var Clients = make(map[*ClientBar]bool)
var Broadcast = make(chan *Message)
var BroadcastReceiver = make(chan *Message)

func receiver(client *ClientBar) {
	for {
		_, p, err := client.Conn.ReadMessage()

		if err != nil {
			delete(Clients, client)

			err = client.Conn.Close()
			if err != nil {
				ErrorLogger.Println(err.Error())
				return
			}
			ErrorLogger.Println("Error read message: " + err.Error())
			return
		}

		var m Message

		err = json.Unmarshal(p, &m)
		if err != nil {
			ErrorLogger.Println(err.Error())
			return
		}

		m.Client = client

		if m.Type == "createOrder" {
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
					delete(Clients, client)
					err := client.Conn.Close()
					if err != nil {
						return
					}
					ErrorLogger.Println("Error write message: " + err.Error())
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
}
