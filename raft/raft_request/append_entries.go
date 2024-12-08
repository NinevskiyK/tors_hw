package raft_request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	state "raft/raft_state"
)

type AppendEntriesReq struct {
	Term         int         `json:"term"`
	LeaderId     int         `json:"leader_id"`
	PrevLogIndex int         `json:"prev_log_index"`
	PrevLogTerm  int         `json:"prev_log_term"`
	Entries      []state.Log `json:"entries"`
	LeaderCommit int         `json:"leader_commit"`
}

type AppendEntriesResp struct {
	Term    int  `json:"term"`
	Success bool `json:"success"`
}

func CreateAppendEntries(index int) (AppendEntriesReq, int, error) {
	state.State.Mu.Lock()
	defer state.State.Mu.Unlock()

	if state.State.Role != state.Leader {
		return AppendEntriesReq{}, 0, errors.New("not a leader")
	}
	last_log_index := state.State.VolatileLeaderState.NextIndex[index]
	last_log_term := 0
	if last_log_index > 0 {
		last_log_term = state.State.PersistentState.Log[last_log_index-1].Term
	}
	reqBody := AppendEntriesReq{
		Term:         state.State.PersistentState.CurrentTerm,
		LeaderId:     state.Me,
		PrevLogIndex: last_log_index,
		PrevLogTerm:  last_log_term,
		Entries:      state.State.PersistentState.Log[last_log_index:],
		LeaderCommit: state.State.VolatileState.CommitIndex,
	}

	return reqBody, len(state.State.PersistentState.Log), nil
}

func AppendEntries(index int, addr string) (AppendEntriesResp, int, error) {
	reqBody, sz, err := CreateAppendEntries(index)
	if err != nil {
		return AppendEntriesResp{}, 0, err
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("failed to marshal request: %s", err)
	}

	client := http.Client{Timeout: time.Millisecond * (state.HeartbeatTimeoutMin / 8)}

	url := fmt.Sprintf("http://%s/append-entries", addr)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// log.Printf("failed to send POST request: %s", err)
		return AppendEntriesResp{Term: -1, Success: false}, 0, nil
	}
	defer resp.Body.Close()

	var resBody AppendEntriesResp
	if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
		log.Fatalf("failed to decode response: %s", err)
	}

	// log.Printf("Received AppendEntriesResp: %+v\n", resBody)
	return resBody, sz, nil
}
