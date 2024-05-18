package main

import (
	"fervexo/hardwire/server"
)

func main() {
	s := server.CreateServer()

	s.AddResource("resource1")
	s.AddResource("resource2")

	s.AddCommand("list", server.ListResources)
	s.AddCommand("lock", server.LockResource)

	s.Run(":8080")
}
