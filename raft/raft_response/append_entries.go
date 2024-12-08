package raft_response

import (
	"encoding/json"
	"net/http"

	"raft/raft_main"
	rpc "raft/raft_request"
	state "raft/raft_state"
)

func HandleAppendEntriesReq(req rpc.AppendEntriesReq) rpc.AppendEntriesResp {
	state.HeartbeatChan <- struct{}{}

	state.State.Mu.Lock()
	defer state.State.Mu.Unlock()

	if req.Term < state.State.PersistentState.CurrentTerm {
		return rpc.AppendEntriesResp{
			Term:    state.State.PersistentState.CurrentTerm,
			Success: false,
		}
	}

	if req.Term > state.State.PersistentState.CurrentTerm {
		raft_main.ConvertToFollower(req.Term)
	}

	state.LeaderNum = req.LeaderId
	if len(state.State.PersistentState.Log) < req.PrevLogIndex ||
		req.PrevLogIndex > 0 && state.State.PersistentState.Log[req.PrevLogIndex-1].Term != req.PrevLogTerm {
		return rpc.AppendEntriesResp{
			Term:    state.State.PersistentState.CurrentTerm,
			Success: false,
		}
	}

	state.State.PersistentState.Log = append(state.State.PersistentState.Log[:req.PrevLogIndex], req.Entries...)
	state.WritePersState(state.State.PersistentState)
	if req.LeaderCommit >= state.State.VolatileState.CommitIndex {
		raft_main.Commit(req.LeaderCommit)
	}

	return rpc.AppendEntriesResp{
		Term:    state.State.PersistentState.CurrentTerm,
		Success: true,
	}
}

func AppendEntriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var reqBody rpc.AppendEntriesReq
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Failed to decode JSON request", http.StatusBadRequest)
		return
	}

	// log.Printf("Received AppendEntriesReq: %+v\n", reqBody)
	resBody := HandleAppendEntriesReq(reqBody)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resBody); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}
