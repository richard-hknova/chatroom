package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func (s *Server) websocketHandler(c *gin.Context) {
	conn, err := s.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Error when establishing server:", err)
		return
	}
	defer conn.Close()
	username := c.Query("username")
	s.Clients[username] = conn
	fmt.Println("user " + username + " is now online.")
}
