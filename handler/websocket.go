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
		fmt.Println("Error establishing server:", err)
		return
	}
	defer conn.Close()
	s.Clients[username] = conn
	fmt.Println("user " + username + " is now online.")
	s.handleMsg(conn, username)
	s.handleOffline(username)
}

func (s *Server) handleMsg(conn *websocket.Conn, username string) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		msg := database.Message{}
		json.Unmarshal(message, &msg)
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
		}
		s.DB.AddMsg(username, msg)
	}
}

func (s *Server) handleOffline(username string) {
	delete(s.Clients, username)
	friends, err := s.DB.GetFriends(username)
	if err != nil {
		fmt.Println("Error getting friend list:", err)
		return
	}
	for _, friend := range friends {
		if client, found := s.Clients[friend.Username]; found {
			data, err := json.Marshal(WebSocketMsg{"Offline Alert", username})
			if err != nil {
				fmt.Println("Error packing message:", err)
				continue
			}
			client.WriteMessage(websocket.TextMessage, data)
		}
	}
	fmt.Println("user " + username + " is now offline.")
}
