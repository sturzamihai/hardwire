package main

import (
	"fervexo/hardwire/server"
)

func main() {
	s := server.CreateServer()

	s.AddResource("Company car")
	s.AddResource("Meeting room")
	s.AddResource("Workstation")

	s.AddCommand("help", server.CommandHelp)

	s.AddCommand("list", server.ListResources)
	s.AddHelpCommand("list", server.ListResourcesHelp)

	s.AddCommand("lock", server.LockResource)
	s.AddHelpCommand("lock", server.LockResourceHelp)

	s.AddCommand("unlock", server.UnlockResource)
	s.AddHelpCommand("unlock", server.UnlockResourceHelp)

	s.AddCommand("reserve", server.ReserveResource)
	s.AddHelpCommand("reserve", server.ReserveResourceHelp)

	s.AddCommand("update", server.UpdateReservation)
	s.AddHelpCommand("update", server.UpdateReservationHelp)

	s.AddCommand("cancel", server.CancelReservation)
	s.AddHelpCommand("cancel", server.CancelReservationHelp)

	s.Run(":8080")
}
