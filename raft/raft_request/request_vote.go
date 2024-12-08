package raft_request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	state "raft/raft_state"
	"time"
)

type RequestVoteReq struct {
	Term         int `json:"term"`
	CandidateId  int `json:"candidate_id"`
	LastLogIndex int `json:"last_log_index"`
	LastLogTerm  int `json:"last_log_term"`
}

type RequestVoteRes struct {
	Term        int  `json:"term"`
	VoteGranted bool `json:"vote_granted"`
}

func CreateRequestVote() (RequestVoteReq, error) {
	state.State.Mu.Lock()
	defer state.State.Mu.Unlock()
	if state.State.Role != state.Candidate {
		return RequestVoteReq{}, errors.New("not a candidate")
	}

	last_log_index := len(state.State.PersistentState.Log)
	last_log_term := 0
	if last_log_index > 0 {
		last_log_term = state.State.PersistentState.Log[last_log_index-1].Term
	}

	reqBody := RequestVoteReq{
		Term:         state.State.PersistentState.CurrentTerm,
		CandidateId:  state.Me,
		LastLogIndex: last_log_index,
		LastLogTerm:  last_log_term,
	}

	return reqBody, nil
}
func RequestVote(addr string) (RequestVoteRes, error) {
	reqBody, err := CreateRequestVote()
	if err != nil {
		return RequestVoteRes{}, err
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("failed to marshal request: %s", err)
	}

	client := http.Client{Timeout: time.Millisecond * (state.HeartbeatTimeoutMin / 8)}

	url := fmt.Sprintf("http://%s/request-vote", addr)

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// log.Printf("failed to send POST request: %s", err)
		return RequestVoteRes{Term: -1, VoteGranted: false}, nil
	}
	defer resp.Body.Close()

	var resBody RequestVoteRes
	if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
		log.Fatalf("failed to decode response: %s", err)
	}

	// log.Printf("Received ReqVote Response: %+v\n", resBody)

	return resBody, nil
}
