package main

import (
	"fervexo/hardwire/server"
)

func main() {
	s := server.CreateServer()

	s.AddResource("Company car")
	s.AddResource("Meeting room")
	s.AddResource("Workstation")

	s.AddCommand("list", server.ListResources)
	s.AddCommand("lock", server.LockResource)
	s.AddCommand("unlock", server.UnlockResource)
	s.AddCommand("reserve", server.ReserveResource)
	s.AddCommand("update", server.UpdateReservation)
	s.AddCommand("cancel", server.CancelReservation)

	s.Run(":8080")
}
