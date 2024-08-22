package main

import "chatroom/handler"

func main() {
	server := handler.NewServer()
	server.Start()
}
