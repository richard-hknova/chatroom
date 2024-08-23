package handler

import (
	"chatroom/database"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketMsg struct {
	MsgType string
	Content any
}

func (s *Server) websocketHandler(c *gin.Context) {
	username := c.GetString("username")
	conn, err := s.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Error when establishing server:", err)
		return
	}
	defer conn.Close()
	s.Clients[username] = conn
	fmt.Println("user " + username + " is now online.")
	s.handleMsg(conn, username)
	delete(s.Clients, username)
	fmt.Println("user " + username + " is now offline.")
}

func (s *Server) handleMsg(conn *websocket.Conn, username string) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		msg := &database.Message{}
		json.Unmarshal(message, msg)
		receiver, found := s.Clients[msg.Receiver]
		msg.Sender = username
		if found {
			msg.Received = true
			data, err := json.Marshal(WebSocketMsg{"Message", msg})
			if err != nil {
				fmt.Println("Error packing message:", err)
				return
			}
			receiver.WriteMessage(websocket.TextMessage, data)
		} // Else store messages into database
	}
}
