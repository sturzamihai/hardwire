package server

import (
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

func CommandHelp(s *Server, c *Client, msg string) *map[string]interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	args := strings.Split(msg, " ")
	if len(args) < 2 {

		command_list := []string{}
		for command := range s.Commands {
			if strings.HasPrefix(command, "help_") {
				command_list = append(command_list, strings.TrimPrefix(command, "help_"))
			}
		}

		return &map[string]interface{}{"commands": command_list}
	}

	command := args[1]
	helpCommand, exists := s.Commands["help_"+command]

	if !exists {
		return &map[string]interface{}{"error": "Command does not exist"}
	}

	return helpCommand(s, c, "")
}

func ListResources(s *Server, c *Client, msg string) *map[string]interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	resources := []*Resource{}
	for _, resource := range s.Resources {
		resources = append(resources, resource)
	}

	log.Println("Listed resources for client: ", c.Name)

	return &map[string]interface{}{"resources": resources}
}

func ListResourcesHelp(s *Server, c *Client, msg string) *map[string]interface{} {
	return &map[string]interface{}{"message": "List all resources. Usage: list"}
}

func LockResource(s *Server, c *Client, msg string) *map[string]interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	args := strings.Split(msg, " ")
	if len(args) != 2 {
		return &map[string]interface{}{"error": "Invalid number of arguments"}
	}

	resourceId := args[1]
	uuid, err := uuid.Parse(resourceId)

	if err != nil {
		return &map[string]interface{}{"error": "Invalid resource ID"}
	}

	resource, exists := s.Resources[uuid]

	if !exists {
		return &map[string]interface{}{"error": "Resource does not exist"}
	}

	reservation, success := resource.Lock(c)
	if !success {
		return &map[string]interface{}{"error": "Resource is already locked"}
	}

	log.Println("Locked resource for client: ", c.Name)
	s.Broadcast("Resource " + resource.Name + " has been locked by " + c.Name + " until " + reservation.End.String())
	return &map[string]interface{}{"reservation": reservation}
}

func LockResourceHelp(s *Server, c *Client, msg string) *map[string]interface{} {
	return &map[string]interface{}{"message": "Lock a resource. Usage: lock [resource_id]"}
}

func UnlockResource(s *Server, c *Client, msg string) *map[string]interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	args := strings.Split(msg, " ")
	if len(args) != 2 {
		return &map[string]interface{}{"error": "Invalid number of arguments"}
	}

	resourceId := args[1]
	uuid, err := uuid.Parse(resourceId)

	if err != nil {
		return &map[string]interface{}{"error": "Invalid resource ID"}
	}

	resource, exists := s.Resources[uuid]

	if !exists {
		return &map[string]interface{}{"error": "Resource does not exist"}
	}

	success := resource.Unlock(c)
	if !success {
		return &map[string]interface{}{"error": "Resource is not locked by client"}
	}

	log.Println("Unlocked resource for client: ", c.Name)
	s.Broadcast("Resource " + resource.Name + " has been unlocked. It is now available for reservation.")
	return &map[string]interface{}{"message": "Successfully unlocked resource"}
}

func UnlockResourceHelp(s *Server, c *Client, msg string) *map[string]interface{} {
	return &map[string]interface{}{"message": "Unlock a resource. Usage: unlock [resource_id]"}
}

func ReserveResource(s *Server, c *Client, msg string) *map[string]interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	args := strings.Split(msg, " ")
	if len(args) != 5 {
		return &map[string]interface{}{"error": "Invalid number of arguments"}
	}

	resourceId := args[1]
	resourceUuid, err := uuid.Parse(resourceId)

	if err != nil {
		return &map[string]interface{}{"error": "Invalid resource ID"}
	}

	resource, exists := s.Resources[resourceUuid]

	if !exists {
		return &map[string]interface{}{"error": "Resource does not exist"}
	}

	reservationId := args[2]
	reservationUuid, err := uuid.Parse(reservationId)

	if err != nil {
		return &map[string]interface{}{"error": "Invalid reservation ID"}
	}

	start, err := time.Parse(time.DateOnly, args[3])

	if err != nil {
		return &map[string]interface{}{"error": "Invalid start time. Format should be YYYY-MM-DD"}
	}

	end, err := time.Parse(time.DateOnly, args[4])

	if err != nil {
		return &map[string]interface{}{"error": "Invalid end time. Format should be YYYY-MM-DD"}
	}

	reservation, success := resource.Reserve(c, reservationUuid, start, end)

	if !success {
		return &map[string]interface{}{"error": "Resource is already reserved or not locked by client"}
	}

	log.Println("Reserved resource for client: ", c.Name)
	s.Broadcast("Resource " + resource.Name + " has been reserved by " + c.Name + " from " + start.String() + " to " + end.String())

	return &map[string]interface{}{"reservation": reservation}
}

func ReserveResourceHelp(s *Server, c *Client, msg string) *map[string]interface{} {
	return &map[string]interface{}{"message": "Reserve a resource. Usage: reserve [resource_id] [reservation_id] [start|YYYY-MM-DD] [end|YYYY-MM-DD]"}
}

func UpdateReservation(s *Server, c *Client, msg string) *map[string]interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	args := strings.Split(msg, " ")
	if len(args) != 5 {
		return &map[string]interface{}{"error": "Invalid number of arguments"}
	}

	resourceId := args[1]
	resourceUuid, err := uuid.Parse(resourceId)

	if err != nil {
		return &map[string]interface{}{"error": "Invalid resource ID"}
	}

	resource, exists := s.Resources[resourceUuid]

	if !exists {
		return &map[string]interface{}{"error": "Resource does not exist"}
	}

	reservationId := args[2]
	reservationUuid, err := uuid.Parse(reservationId)

	if err != nil {
		return &map[string]interface{}{"error": "Invalid reservation ID"}
	}

	start, err := time.Parse(time.DateOnly, args[3])

	if err != nil {
		return &map[string]interface{}{"error": "Invalid start time. Format should be YYYY-MM-DD"}
	}

	end, err := time.Parse(time.DateOnly, args[4])

	if err != nil {
		return &map[string]interface{}{"error": "Invalid end time. Format should be YYYY-MM-DD"}
	}

	success := resource.UpdateReservation(c, reservationUuid, start, end)

	if !success {
		return &map[string]interface{}{"error": "Resource is not reserved by client or reservation does not exist"}
	}

	log.Println("Updated reservation for client: ", c.Name)
	s.Broadcast("Reservation for resource " + resource.Name + " has been updated by " + c.Name + " from " + start.String() + " to " + end.String())

	return &map[string]interface{}{"message": "Successfully updated reservation"}
}

func UpdateReservationHelp(s *Server, c *Client, msg string) *map[string]interface{} {
	return &map[string]interface{}{"message": "Update a reservation. Usage: update [resource_id] [reservation_id] [start|YYYY-MM-DD] [end|YYYY-MM-DD]"}
}

func CancelReservation(s *Server, c *Client, msg string) *map[string]interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	args := strings.Split(msg, " ")
	if len(args) != 3 {
		return &map[string]interface{}{"error": "Invalid number of arguments"}
	}

	resourceId := args[1]
	resourceUuid, err := uuid.Parse(resourceId)

	if err != nil {
		return &map[string]interface{}{"error": "Invalid resource ID"}
	}

	resource, exists := s.Resources[resourceUuid]

	if !exists {
		return &map[string]interface{}{"error": "Resource does not exist"}
	}

	reservationId := args[2]
	reservationUuid, err := uuid.Parse(reservationId)

	if err != nil {
		return &map[string]interface{}{"error": "Invalid reservation ID"}
	}

	success := resource.CancelReservation(c, reservationUuid)

	if !success {
		return &map[string]interface{}{"error": "Resource is not reserved by client or reservation does not exist"}
	}

	log.Println("Cancelled reservation for client: ", c.Name)
	s.Broadcast("Reservation for resource " + resource.Name + " has been cancelled by " + c.Name)

	return &map[string]interface{}{"message": "Successfully cancelled reservation"}
}

func CancelReservationHelp(s *Server, c *Client, msg string) *map[string]interface{} {
	return &map[string]interface{}{"message": "Cancel a reservation. Usage: cancel [resource_id] [reservation_id]"}
}
