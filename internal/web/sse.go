package web

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Hub is the SSE fan-out. Two kinds of message exist, matching the vendored
// htmx 4 hx-sse extension:
//
//   - unnamed (`data:` only)  → swapped automatically into the connecting
//     element (that is what replaced `sse-swap`, which is removed in v4)
//   - named   (`event: x`)    → dispatched as a DOM event on that element
//
// Each stream is keyed by an opaque id so the roadmap pump can feed one page
// without broadcasting to every dashboard.
type Hub struct {
	mu      sync.Mutex
	clients map[*stream]struct{}
}

type stream struct {
	id string
	ch chan string
}

func NewHub() *Hub { return &Hub{clients: map[*stream]struct{}{}} }

func (h *Hub) Subscribe(id string) *stream {
	s := &stream{id: id, ch: make(chan string, 32)}
	h.mu.Lock()
	h.clients[s] = struct{}{}
	h.mu.Unlock()
	return s
}

func (h *Hub) Unsubscribe(s *stream) {
	h.mu.Lock()
	if _, ok := h.clients[s]; ok {
		delete(h.clients, s)
		close(s.ch)
	}
	h.mu.Unlock()
}

// Send delivers an unnamed message (auto-swapped into the connecting element).
func (h *Hub) Send(id, html string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for s := range h.clients {
		if s.id != id {
			continue
		}
		select {
		case s.ch <- "data: " + oneLine(html) + "\n\n":
		default:
		}
	}
}

// Event delivers a named event, which the browser sees as a DOM event.
func (h *Hub) Event(id, name, data string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for s := range h.clients {
		if s.id != id {
			continue
		}
		select {
		case s.ch <- fmt.Sprintf("event: %s\ndata: %s\n\n", name, oneLine(data)):
		default:
		}
	}
}

// Serve streams to one subscriber. Unnamed messages arrive on ch already
// formatted as SSE frames.
func (h *Hub) Serve(w http.ResponseWriter, r *http.Request, id, initial string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	if initial != "" {
		if _, err := fmt.Fprint(w, "data: "+oneLine(initial)+"\n\n"); err != nil {
			return
		}
		flusher.Flush()
	}

	s := h.Subscribe(id)
	defer h.Unsubscribe(s)

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg, open := <-s.ch:
			if !open {
				return
			}
			if _, err := fmt.Fprint(w, msg); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// oneLine collapses a fragment to a single SSE `data:` line. Multi-line data is
// legal SSE only as repeated data: fields, which htmx would concatenate, so we
// keep fragments on one line and let the browser parse the HTML.
func oneLine(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\n':
			b = append(b, ' ')
		case '\r':
			// drop
		default:
			b = append(b, s[i])
		}
	}
	return string(b)
}