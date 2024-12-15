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
	internal_err, res := raft_main.AddToLog(raft_state.Log{Op: "Update", Uuid: uuid, Body: string(body[:]), Term: 0})
	if !internal_err {
		if res {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Resource %s updated with %s\n", uuid, body)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "Resourse %s doens't exists", uuid)
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "try again")
	}
}
