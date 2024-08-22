package handler

import (
	"chatroom/database"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

type Server struct {
	DB       *database.DB
	Upgrader *websocket.Upgrader
	Secret   string
	Clients  map[string]*websocket.Conn
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
	return &Server{DB: DB, Upgrader: Upgrader, Secret: os.Getenv("SECRET_KEY")}
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		fmt.Println(tokenString)
		splitToken := strings.Split(tokenString, " ")
		if splitToken[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
			c.Abort()
			return
		}
		if splitToken[1] == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(splitToken[1], &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.Secret), nil
		})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims := token.Claims.(*CustomClaims)
		c.Set("username", claims.Username)
		c.Set("avatar", claims.Avatar)
		c.Next()
	}
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
	app.GET("/ws", s.authMiddleware(), s.websocketHandler)
	app.POST("/user/signin", s.signInHandler)
	app.POST("/user/signup", s.signUpHandler)
	app.GET("/user/search", s.authMiddleware(), s.searchUserHandler)
	app.POST("/friend/request", s.authMiddleware(), s.requestFriendHandler)
	app.PUT("/friend/accept", s.authMiddleware(), s.acceptFriendHandler)
	app.DELETE("/friend/delete", s.authMiddleware(), s.deleteFriendHandler)
	app.Run(":8080")
}
