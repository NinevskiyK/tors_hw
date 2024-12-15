package rust_wrapper

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"raft/raft_main"
	"raft/raft_response"
	state "raft/raft_state"
	"sync"
	"time"
)

func HeartBeatGetter(wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		ctx, cancel := context.WithCancel(context.Background())
		heartbeatTimeout := rand.IntN(state.HeartbeatTimeoutMax-state.HeartbeatTimeoutMin) + state.HeartbeatTimeoutMin
		timer := time.AfterFunc(time.Duration(heartbeatTimeout)*time.Millisecond, func() {
			cancel()
		})

		select {
		case <-state.HeartbeatChan:
			// log.Println("Heartbeat received, resetting timer")
			timer.Stop()
			cancel()
		case <-ctx.Done():
			go raft_main.StartElection()
		}
	}
}

func ServeRPC(wg *sync.WaitGroup) {
	defer wg.Done()
	ip := os.Getenv("IP")
	port := os.Getenv("RAFT_PORT")

	address := fmt.Sprintf("%s:%s", ip, port)

	http.HandleFunc("/request-vote", raft_response.RequestVoteHandler)
	http.HandleFunc("/append-entries", raft_response.AppendEntriesHandler)
	http.HandleFunc("/request-read/", raft_response.ReadHandler)

	log.Printf("Starting raft RPC Server at http://%s\n", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}

func Serve() {
	state.InitState()

	var wg sync.WaitGroup
	wg.Add(1)
	go ServeRPC(&wg)

	time.Sleep(1 * time.Second)

	wg.Add(1)
	state.HeartbeatChan = make(chan struct{})
	go HeartBeatGetter(&wg)

	wg.Wait()
}
