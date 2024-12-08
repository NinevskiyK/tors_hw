package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	rust_wrapper "raft/raft_wrapper"
	"raft/request_handler"
	statemachine "raft/state_machine"
)

func main() {
	ip := os.Getenv("IP")
	port := os.Getenv("HTTP_PORT")

	address := fmt.Sprintf("%s:%s", ip, port)

	http.HandleFunc("/", request_handler.Router)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Starting server at http://%s\n", address)
		if err := http.ListenAndServe(address, nil); err != nil {
			log.Fatalf("Server failed: %s", err)
		}
	}()
	go func() {
		defer wg.Done()
		statemachine.Init()
		rust_wrapper.Serve()
	}()
	wg.Wait()
	log.Printf("Service stopped\n")
}
