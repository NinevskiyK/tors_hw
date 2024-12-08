package request_handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"raft/raft_main"
	"raft/raft_state"
	state "raft/raft_state"
)

func UpdateHandler(w http.ResponseWriter, r *http.Request, uuid string) {
	log.Printf("UPDATE: Resource with UUID %s\n", uuid)
	if ok, addr := state.CheckIfLeader(); !ok {
		w.WriteHeader(http.StatusFound)
		fmt.Fprintf(w, "Contact leader: %s", addr)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal err: %s", body)
		return
	}
	if raft_main.AddToLog(raft_state.Log{Op: "Update", Uuid: uuid, Body: string(body[:]), Term: 0}) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Resource %s deleted\n", uuid)
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Resource %s created with body %s\n", uuid, body)
	}
}
