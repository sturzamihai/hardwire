package client

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
}

func CreateClient(addr string, name string) (*Client, error) {
	headers := make(map[string][]string)
	headers["Authorization"] = []string{"Name " + name}
	conn, _, err := websocket.DefaultDialer.Dial(addr, headers)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

func (c *Client) SendMessage(message string) error {
	return c.conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func (c *Client) Listen() {
	log.Println("Listening for messages...")
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		log.Println("Message received: ", string(msg))
	}
}

func (c *Client) Close() {
	c.conn.Close()
}
