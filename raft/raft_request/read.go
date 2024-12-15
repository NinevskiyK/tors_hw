package raft_request

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	state "raft/raft_state"
	"time"
)

type RequestReadResp struct {
	Value  string `json:"value"`
	Exists bool   `json:"exists"`
}

func RequestRead(addr, key string) (RequestReadResp, error) {
	url := fmt.Sprintf("http://%s/request-read/%s", addr, key)

	client := http.Client{Timeout: time.Millisecond * (state.HeartbeatTimeoutMin / 8)}
	resp, err := client.Get(url)
	if err != nil {
		// log.Printf("failed to send POST request: %s", err)
		return RequestReadResp{}, err
	}
	defer resp.Body.Close()

	var resBody RequestReadResp
	if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
		log.Fatalf("failed to decode response: %s", err)
	}

	return resBody, nil
}
