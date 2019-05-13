package http

import (
	"fmt"
	"net/http"
	"time"
)

var up = time.Now()

func StatusHandler(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)

	var status = struct {
		Status string        `json:"status"`
		Start  time.Time     `json:"started"`
		Uptime time.Duration `json:"uptime"`
	}{
		Status: "ready",
		Start:  up,
		Uptime: time.Duration(time.Since(up)),
	}

	encoded := json.NewEncoder(rw)
	encoded.SetIndent("", "	")
	err := encoded.Encode(status)
	if err != nil {
		fmt.Fprintf(rw, "Error encoding status data: %v\n", err)
	}
}

func HealthHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	fmt.Fprintln(rw, `ok`)

}
