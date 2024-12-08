package request_handler

import (
	"fmt"
	"log"
	"net/http"
	state "raft/raft_state"
	statemachine "raft/state_machine"
)

func ReadHandler(w http.ResponseWriter, r *http.Request, uuid string) {
	log.Printf("READ: Resource with UUID %s\n", uuid)
	if ok, addr := state.CheckIfLeader(); !ok {
		w.WriteHeader(http.StatusFound)
		fmt.Fprintf(w, "Contact leader: %s", addr)
		return
	}
	if value, err := statemachine.Read(uuid); err == nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s", value)
		return
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "%s", err.Error())
	}
}
