package request_handler

import (
	"fmt"
	"log"
	"net/http"
	"raft/raft_main"
)

func ReadHandler(w http.ResponseWriter, r *http.Request, uuid string) {
	log.Printf("READ: Resource with UUID %s\n", uuid)
	value, not_stale := raft_main.GetQuorumRead(uuid)
	if not_stale {
		if value.Exists {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%s", value.Value)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "value not found")
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "I'm a stale replica")
	}
}
