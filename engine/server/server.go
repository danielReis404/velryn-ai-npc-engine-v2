package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"ai-agent-engine/engine"
)

func Serve(addr string, world *engine.World) error {
	mux := http.NewServeMux()

	registerAPIRoutes(mux, world)

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ch := world.Subscribe()
		world.ViewerConnected()
		log.Println("👀 viewer connected — the world starts ticking (or keeps ticking) for them")

		defer func() {
			world.ViewerDisconnected()
			world.Unsubscribe(ch)
			log.Println("👋 viewer disconnected")
		}()

		ping := time.NewTicker(10 * time.Second)
		defer ping.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ping.C:
				if _, err := fmt.Fprint(w, ": keepalive\n\n"); err != nil {
					return
				}
				flusher.Flush()
			case data, open := <-ch:
				if !open {
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	})

	mux.Handle("/", http.FileServer(http.Dir("server/static")))

	log.Printf("🌍 serving Velryn on http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}
