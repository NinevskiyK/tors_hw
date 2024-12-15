package raft_response

import (
	"encoding/json"
	"net/http"
	"raft/raft_request"
	statemachine "raft/state_machine"
	"strings"
)

func ReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	uuid := strings.Split(r.URL.Path, "/")[len(strings.Split(r.URL.Path, "/"))-1]
	if uuid == "" {
		http.Error(w, "UUID is required", http.StatusBadRequest)
		return
	}

	value, err := statemachine.Read(uuid)
	var resBody raft_request.RequestReadResp
	if err != nil {
		resBody = raft_request.RequestReadResp{Value: "", Exists: false}
	} else {
		resBody = raft_request.RequestReadResp{Value: value, Exists: true}
	}
	// log.Printf("%+v", resBody)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resBody); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}
