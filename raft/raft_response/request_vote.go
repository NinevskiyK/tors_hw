package raft_response

import (
	"encoding/json"
	"net/http"

	"raft/raft_main"
	rpc "raft/raft_request"
	state "raft/raft_state"
)

func HandleRequestVoteReq(req rpc.RequestVoteReq) rpc.RequestVoteRes {
	state.State.Mu.Lock()
	defer state.State.Mu.Unlock()

	voted := false
	if req.Term > state.State.PersistentState.CurrentTerm {
		if state.State.PersistentState.VotedFor == -1 || state.State.PersistentState.VotedFor == req.CandidateId {
			my_last_index := len(state.State.PersistentState.Log)
			my_last_term := 0
			if my_last_index > 0 {
				my_last_term = state.State.PersistentState.Log[my_last_index-1].Term
			}

			if my_last_term < req.LastLogTerm || (my_last_term == req.LastLogTerm && my_last_index <= req.LastLogIndex) {
				raft_main.ConvertToFollower(req.Term)
				state.State.PersistentState.VotedFor = req.CandidateId
				state.WritePersState(state.State.PersistentState)
				voted = true
			}
		}
	}

	resBody := rpc.RequestVoteRes{
		Term:        state.State.PersistentState.CurrentTerm,
		VoteGranted: voted,
	}

	return resBody
}

func RequestVoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var reqBody rpc.RequestVoteReq
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Failed to decode JSON request", http.StatusBadRequest)
		return
	}

	// log.Printf("Received RequestVote: %+v\n", reqBody)

	resBody := HandleRequestVoteReq(reqBody)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resBody); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}
