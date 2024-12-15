package request_handler

import (
	"net/http"
	"os"
	"strings"
)

func Router(w http.ResponseWriter, r *http.Request) {
	uuid := strings.TrimPrefix(r.URL.Path, "/")
	if uuid == "" {
		http.Error(w, "UUID is required", http.StatusBadRequest)
		os.Exit(0)
		return
	}

	switch r.Method {
	case http.MethodPost:
		CreateHandler(w, r, uuid)
	case http.MethodGet:
		ReadHandler(w, r, uuid)
	case http.MethodPut:
		UpdateHandler(w, r, uuid)
	case http.MethodPatch:
		CasHandler(w, r, uuid)
	case http.MethodDelete:
		DeleteHandler(w, r, uuid)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
