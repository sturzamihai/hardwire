package server

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Name string          `json:"name"`
	Conn *websocket.Conn `json:"-"`
}

func (c *Client) SendMessage(message string) {
	err := c.Conn.WriteMessage(websocket.TextMessage, []byte(message))

	if err != nil {
		log.Println("Error sending message to client: ", err)
	}
}

func (c *Client) Close() {
	c.Conn.Close()
}
