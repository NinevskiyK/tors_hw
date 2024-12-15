package request_handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"raft/raft_main"
	state "raft/raft_state"
	"strings"
)

func CasHandler(w http.ResponseWriter, r *http.Request, uuid string) {
	log.Printf("CAS: Resource with UUID %s\n", uuid)
	if ok, addr := state.CheckIfLeader(); !ok {
		w.Header().Set("Location", fmt.Sprintf("http://127.0.0.1:%s/%s", strings.Split(addr, ":")[1], uuid))
		w.WriteHeader(http.StatusTemporaryRedirect)
		log.Printf("Contact leader: %s", addr)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal err: %s", body)
		return
	}
	internal_err, res := raft_main.AddToLog(state.Log{Op: "CAS", Uuid: uuid, Body: string(body[:]), Term: 0})
	if !internal_err {
		if res {
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprintf(w, "CAS ok, %s, %s", uuid, body)
		} else {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprintf(w, "CAS didn't succeed, %s, %s", uuid, body)
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Try again")
	}
}
