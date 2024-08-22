package handler

import (
	"chatroom/database"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func getAuthFromHeader(c *gin.Context) (string, string, error) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		return "", "", errors.New("invalid header")
	}
	encodedCreds := strings.TrimPrefix(authHeader, "Basic ")
	decodedCreds, err := base64.StdEncoding.DecodeString(encodedCreds)
	if err != nil {
		return "", "", errors.New("error decoding header")
	}
	credentials := strings.Split(string(decodedCreds), ":")
	if len(credentials) != 2 {
		return "", "", errors.New("invalid credentials")
	}
	return credentials[0], credentials[1], nil
}

type CustomClaims struct {
	Username string `json:"username"`
	Avatar   int    `json:"avatar"`
	jwt.RegisteredClaims
}

func (s *Server) genToken(user *database.User) (string, error) {
	claims := CustomClaims{
		user.Username,
		user.Avatar,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Secret))
}

func (s *Server) signInHandler(c *gin.Context) {
	username, password, err := getAuthFromHeader(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := s.DB.AuthUser(username, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	requests, err := s.DB.GetRequests(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	friends, err := s.DB.GetFriends(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	token, err := s.genToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type Response struct {
		Profile  database.User
		Requests []database.User
		Friends  []database.User
		Token    string
	}
	response := Response{
		Profile:  database.User{Username: user.Username, Avatar: user.Avatar},
		Requests: requests,
		Friends:  friends,
		Token:    token,
	}
	c.JSON(http.StatusOK, response)
}
func (s *Server) signUpHandler(c *gin.Context) {
	username, password, err := getAuthFromHeader(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := s.DB.GetUser(username)
	if err != nil && err.Error() != "record not found" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if user != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "username already exist"})
		return
	}
	err = s.DB.SetUser(username, password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (s *Server) searchUserHandler(c *gin.Context) {
	username := c.Query("search")
	user, err := s.DB.GetUser(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong. Please try again later."})
		return
	}
	c.JSON(http.StatusOK, &database.User{Username: user.Username, Avatar: user.Avatar})
}

func (s *Server) requestFriendHandler(c *gin.Context) {
	target := c.Query("target")
	username := c.GetString("username")
	avatar := c.GetInt("avatar")
	var user database.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	err := s.DB.SetRequest(username, avatar, target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong. Please try again later."})
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (s *Server) acceptFriendHandler(c *gin.Context) {
	username := c.GetString("username")
	avatar := c.GetInt("avatar")
	var target database.User
	if err := c.ShouldBindJSON(&target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	err := s.DB.SetFriend(username, avatar, target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong. Please try again later."})
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (s *Server) deleteFriendHandler(c *gin.Context) {
	username := c.GetString("username")
	target := c.Query("target")
	err := s.DB.DeleteFriendOrRequest(username, target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong. Please try again later."})
		return
	}
	c.JSON(http.StatusOK, nil)
}
