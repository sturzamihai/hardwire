package server

import (
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
	Commands  map[string]func(*Server, *Client, string) *CommandResponse
	mutex     sync.Mutex
	upgrader  websocket.Upgrader
}

func CreateServer() *Server {
	return &Server{Resources: make(map[uuid.UUID]*Resource), Clients: make(map[string]*Client), Commands: make(map[string]func(*Server, *Client, string) *CommandResponse)}
}

func (s *Server) AddResource(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := uuid.New()
	if _, exists := s.Resources[id]; !exists {
		s.Resources[id] = &Resource{ID: id, Name: name, Reservations: make(map[uuid.UUID]*Reservation)}
	}
}

func (s *Server) AddCommand(name string, command func(*Server, *Client, string) *CommandResponse) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.Commands[name]; !exists {
		s.Commands[name] = command
	}
}

func (s *Server) Broadcast(message string) {
	for _, client := range s.Clients {
		if client.Name != "" {
			client.SendMessage(message)
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
			response := &CommandResponse{Error: "Invalid command"}
			jsonResponse, err := response.toJSONString()

			if err != nil {
				log.Println("Error converting response to JSON: ", err)
			}

			c.SendMessage(jsonResponse)

			continue
		}

		response := command(s, c, message)
		jsonResponse, err := response.toJSONString()

		if err != nil {
			log.Println("Error converting response to JSON: ", err)
		}

		c.SendMessage(jsonResponse)
	}
}

func (s *Server) Run(addr string) {
	http.HandleFunc("/ws", s.handleConnection)
	log.Fatal(http.ListenAndServe(addr, nil))
}
