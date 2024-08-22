package handler

import (
	"chatroom/database"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Server struct {
	DB       *database.DB
	Upgrader *websocket.Upgrader
}

func NewServer() *Server {
	DB, err := database.ConnectDB()
	if err != nil {
		log.Fatalln("Error connecting database:", err)
	}
	Upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return &Server{DB: DB, Upgrader: Upgrader}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
		}

		c.Next()
	}
}

func (s *Server) Start() {
	app := gin.Default()
	app.Use(CORSMiddleware())
	app.GET("/ws", s.websocketHandler)
	app.POST("/user/signin", s.signInHandler)
	app.POST("/user/signup", s.signUpHandler)
	app.GET("/user/search", s.searchUserHandler)
	app.POST("/friend/request", s.requestFriendHandler)
	app.PUT("/friend/accept", s.acceptFriendHandler)
	app.DELETE("/friend/delete", s.deleteFriendHandler)
	app.Run(":8080")
}
