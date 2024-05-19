package main

import (
	"bufio"
	"fervexo/hardwire/client"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

func handleInput(ch chan string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			close(ch)
			return
		}
		ch <- s
	}
}

func main() {
	clientName := flag.String("name", "client1", "Client name")
	flag.Parse()

	c, err := client.CreateClient("ws://localhost:8080/", *clientName)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	log.Println("Client connected as:", *clientName, "...")

	go c.Listen()

	ch := make(chan string)
	go handleInput(ch)

stdinloop:
	for {
		select {
		case stdin, ok := <-ch:
			if !ok {
				break stdinloop
			} else {
				c.SendMessage(stdin)
			}
		case <-time.After(1 * time.Second):

		}
	}
	fmt.Println("Done, stdin must be closed")

}
