package request_handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"raft/raft_main"
	state "raft/raft_state"
)

func CreateHandler(w http.ResponseWriter, r *http.Request, uuid string) {
	log.Printf("CREATE: Resource with UUID %s\n", uuid)
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
	if raft_main.AddToLog(state.Log{Op: "Create", Uuid: uuid, Body: string(body[:]), Term: 0}) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Resource %s created with body %s\n", uuid, body)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Try again")
	}
}
