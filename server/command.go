package server

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/google/uuid"
)

type CommandResponse struct {
	Data  string `json:"data"`
	Error string `json:"error"`
}

func (e *CommandResponse) toJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (e *CommandResponse) toJSONString() (string, error) {
	json, err := e.toJSON()
	return string(json), err
}

func ListResources(s *Server, c *Client, msg string) *CommandResponse {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	resources := []*Resource{}
	for _, resource := range s.Resources {
		resources = append(resources, resource)
	}

	json, err := json.Marshal(resources)

	if err != nil {
		log.Println("Error marshalling resources: ", err)
		return &CommandResponse{Error: "Error marshalling resources"}
	}

	log.Println("Listed resources for client: ", c.Name)
	return &CommandResponse{Data: string(json)}
}

func LockResource(s *Server, c *Client, msg string) *CommandResponse {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	args := strings.Split(msg, " ")
	if len(args) != 2 {
		return &CommandResponse{Error: "Invalid number of arguments"}
	}

	resourceId := args[1]
	uuid, err := uuid.Parse(resourceId)

	if err != nil {
		return &CommandResponse{Error: "Invalid resource ID"}
	}

	resource, exists := s.Resources[uuid]

	if !exists {
		return &CommandResponse{Error: "Resource does not exist"}
	}

	reservation, success := resource.Lock(c)
	if !success {
		return &CommandResponse{Error: "Resource is already locked"}
	}

	json, err := json.Marshal(reservation)

	if err != nil {
		log.Println("Error marshalling reservation: ", err)
		return &CommandResponse{Error: "Error marshalling reservation"}
	}

	log.Println("Locked resource for client: ", c.Name)

	return &CommandResponse{Data: string(json)}
}
