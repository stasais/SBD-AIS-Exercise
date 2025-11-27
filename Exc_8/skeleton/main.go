package main

import (
	"time"
)

func main() {
	go func() {
		// todo start server
	}()
	time.Sleep(1 * time.Second)
	// todo start client
	println("Orders complete!")
}
