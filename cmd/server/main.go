package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"workshop-zero/internal/plan"
	"workshop-zero/internal/store"
	"workshop-zero/internal/web"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	data := flag.String("data", "data", "data directory for the JSONL event store")
	flag.Parse()

	if args := flag.Args(); len(args) > 0 && args[0] == "freeze" {
		out := "dist"
		if len(args) > 1 {
			out = args[1]
		}
		if err := web.Freeze(out); err != nil {
			log.Fatalf("freeze: %v", err)
		}
		log.Printf("frozen static site written to %s", out)
		return
	}

	if err := os.MkdirAll(*data, 0o755); err != nil {
		log.Fatal(err)
	}
	st, err := store.Open(filepath.Join(*data, "events.jsonl"))
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	lib, err := plan.LoadLibrary()
	if err != nil {
		log.Fatal(err)
	}
	srv, err := web.New(st, lib)
	if err != nil {
		log.Fatal(err)
	}

	httpSrv := &http.Server{
		Addr:              *addr,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		// SSE connections are long-lived; do not cap WriteTimeout.
		WriteTimeout: 0,
		IdleTimeout:  120 * time.Second,
	}
	fmt.Printf("workshop-zero listening on %s\n", *addr)
	log.Fatal(httpSrv.ListenAndServe())
}