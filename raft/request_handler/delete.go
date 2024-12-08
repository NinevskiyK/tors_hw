package request_handler

import (
	"fmt"
	"log"
	"net/http"
	"raft/raft_main"
	"raft/raft_state"
	state "raft/raft_state"
)

func DeleteHandler(w http.ResponseWriter, r *http.Request, uuid string) {
	log.Printf("DELETE: Resource with UUID %s\n", uuid)
	if ok, addr := state.CheckIfLeader(); !ok {
		w.WriteHeader(http.StatusFound)
		fmt.Fprintf(w, "Contact leader: %s", addr)
		return
	}
	if raft_main.AddToLog(raft_state.Log{Op: "Delete", Uuid: uuid, Body: "", Term: 0}) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Resource %s deleted\n", uuid)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Try again")
	}
}
