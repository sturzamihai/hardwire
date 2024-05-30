package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Server struct {
	Resources map[uuid.UUID]*Resource
	Clients   map[string]*Client
	Commands  map[string]func(*Server, *Client, string) *map[string]interface{}
	mutex     sync.Mutex
	upgrader  websocket.Upgrader
}

func CreateServer() *Server {
	return &Server{Resources: make(map[uuid.UUID]*Resource), Clients: make(map[string]*Client), Commands: make(map[string]func(*Server, *Client, string) *map[string]interface{})}
}

func (s *Server) AddResource(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := uuid.New()
	if _, exists := s.Resources[id]; !exists {
		s.Resources[id] = &Resource{ID: id, Name: name, Reservations: make(map[uuid.UUID]*Reservation)}
	}
}

func (s *Server) AddCommand(name string, command func(*Server, *Client, string) *map[string]interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.Commands[name]; !exists {
		s.Commands[name] = command
	}
}

func (s *Server) AddHelpCommand(name string, command func(*Server, *Client, string) *map[string]interface{}) {
	s.AddCommand("help_"+name, command)
}

func (s *Server) Broadcast(message string) {
	for _, client := range s.Clients {
		if client.Name != "" {
			message := &map[string]interface{}{"broadcast": message}
			jsonMessage, err := json.Marshal(message)

			if err != nil {
				log.Println("Error converting message to JSON: ", err)
			}

			client.SendMessage(string(jsonMessage))
		}
	}
}

func (s *Server) getNameAuth(token []string) (string, bool) {
	if len(token) != 1 {
		return "", false
	}

	split := strings.Split(token[0], " ")

	if len(split) != 2 {
		return "", false
	}

	if split[0] != "Name" {
		return "", false
	}

	return split[1], true
}

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	log.Println("Received connection...")

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket: ", err)
		return
	}

	name, ok := s.getNameAuth(r.Header["Authorization"])

	if !ok {
		log.Println("Invalid authorization header")
		conn.WriteMessage(websocket.TextMessage, []byte("Invalid authorization header"))
		conn.Close()
		return
	}

	log.Println("Client connected: ", name)
	client := &Client{Name: name, Conn: conn}
	s.Clients[client.Name] = client

	go s.handleClient(client)
}

func (s *Server) handleClient(c *Client) {
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message: ", err)
			return
		}

		message := strings.TrimSpace(string(msg))
		prefix := strings.Split(message, " ")[0]

		command, exists := s.Commands[prefix]
		if !exists {
			response := &map[string]interface{}{"error": "Invalid command"}
			jsonResponse, err := json.Marshal(response)

			if err != nil {
				log.Println("Error converting response to JSON: ", err)
			}

			c.SendMessage(string(jsonResponse))

			continue
		}

		response := command(s, c, message)
		jsonResponse, err := json.Marshal(response)

		if err != nil {
			log.Println("Error converting response to JSON: ", err)
		}

		c.SendMessage(string(jsonResponse))
	}
}

func (s *Server) Run(addr string) {
	log.Println("Server running on port", addr)
	http.HandleFunc("/", s.handleConnection)
	log.Fatal(http.ListenAndServe(addr, nil))
}
