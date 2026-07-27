package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	addr = ":8080"

	// readHeaderTimeout bounds how long a client may take to send its request
	// headers, so an idle connection cannot occupy the server indefinitely.
	readHeaderTimeout = 10 * time.Second
)

// TODO: greet the user by name from config
func main() {
	fmt.Println("Hello, World!")

	server := &http.Server{
		Addr:              addr,
		Handler:           newMux(),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		slog.Error("http server stopped", slog.Any("error", err))
		os.Exit(1)
	}
}

// newMux returns a router with every application route registered.
func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	return mux
}

// healthz reports that the server is alive by responding 200 with the body "ok".
//
// The endpoint answers any request method so that liveness probes never fail on
// method negotiation.
func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// A probe that hangs up mid-response leaves nothing worth recovering.
	_, _ = io.WriteString(w, "ok")
}
