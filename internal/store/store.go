package store

import (
	"bufio"
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"

	"workshop-zero/internal/quiz"
)

// Event is one append-only record in the JSONL log.
type Event struct {
	Time time.Time      `json:"time"`
	Kind string         `json:"kind"`
	Path string         `json:"path,omitempty"`
	IP   string         `json:"ip,omitempty"`
	Meta map[string]any `json:"meta,omitempty"`
}

// Store is an append-only JSONL log plus an in-memory index rebuilt on boot.
// No database, no CGO: the free tier's ephemeral disk is enough, and the
// facilitator can export the raw file at any time.
type Store struct {
	mu      sync.Mutex
	f       *os.File
	enc     *json.Encoder
	count   int
	answers map[string][]string
	results map[string]quiz.Result
	done    map[string]map[int]bool
	dist    map[string]int
}

func Open(path string) (*Store, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	s := &Store{
		f:       f,
		enc:     json.NewEncoder(f),
		answers: map[string][]string{},
		results: map[string]quiz.Result{},
		done:    map[string]map[int]bool{},
		dist:    map[string]int{},
	}
	s.rebuild()
	return s, nil
}

// rebuild folds the whole log into the in-memory index. Append-only means this
// is always safe to re-run; a crash can never corrupt the index beyond the last
// fully written line.
func (s *Store) rebuild() {
	if _, err := s.f.Seek(0, 0); err != nil {
		return
	}
	sc := bufio.NewScanner(s.f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		s.count++
		s.index(e)
	}
	_, _ = s.f.Seek(0, 2)
}

func (s *Store) index(e Event) {
	switch e.Kind {
	case "quiz_reset":
		id, _ := e.Meta["quiz"].(string)
		if id != "" {
			delete(s.answers, id)
		}
	case "answer":
		id, _ := e.Meta["quiz"].(string)
		n := asInt(e.Meta["n"])
		ans, _ := e.Meta["answer"].(string)
		if id != "" && ans != "" {
			s.answers[id] = append(s.answers[id], ans)
		}
		_ = n
	case "quiz_result":
		id, _ := e.Meta["id"].(string)
		raw, ok := e.Meta["result"]
		if !ok || id == "" {
			return
		}
		b, err := json.Marshal(raw)
		if err != nil {
			return
		}
		var res quiz.Result
		if err := json.Unmarshal(b, &res); err != nil {
			return
		}
		res.ID = id
		s.results[id] = res
	case "done":
		id, _ := e.Meta["roadmap"].(string)
		w := asInt(e.Meta["week"])
		if id == "" {
			return
		}
		if s.done[id] == nil {
			s.done[id] = map[int]bool{}
		}
		if done, _ := e.Meta["done"].(bool); done {
			s.done[id][w] = true
		} else {
			delete(s.done[id], w)
		}
	}
}

func asInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}

func (s *Store) Append(e Event) error {
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.enc.Encode(e); err != nil {
		return err
	}
	s.count++
	s.index(e)
	return nil
}

func (s *Store) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.count
}

// Reset clears the recorded answers for one attempt. Replaying a quiz id must
// start from zero, otherwise a second pass inherits the first pass's answers.
func (s *Store) Reset(quizID string) {
	_ = s.Append(Event{Kind: "quiz_reset", Meta: map[string]any{"quiz": quizID}})
}

// Answer records one answer for a quiz attempt.
func (s *Store) Answer(quizID string, n int, answer string) {
	_ = s.Append(Event{Kind: "answer", Meta: map[string]any{"quiz": quizID, "n": n, "answer": answer}})
}

func (s *Store) Answers(quizID string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.answers[quizID]...)
}

func (s *Store) SaveResult(res quiz.Result) {
	_ = s.Append(Event{Kind: "quiz_result", Meta: map[string]any{"id": res.ID, "result": res}})
}

func (s *Store) Result(id string) (quiz.Result, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.results[id]
	return r, ok
}

func (s *Store) SetDone(roadmapID string, week int, done bool) {
	_ = s.Append(Event{Kind: "done", Meta: map[string]any{"roadmap": roadmapID, "week": week, "done": done}})
}

func (s *Store) Done(roadmapID string) map[int]bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := map[int]bool{}
	for k, v := range s.done[roadmapID] {
		out[k] = v
	}
	return out
}

// Distribution returns how many roadmaps contain each module, which is the
// number the facilitator dashboard actually needs.
func (s *Store) Distribution() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := map[string]int{}
	for _, r := range s.results {
		for _, m := range r.Selected {
			out[m]++
		}
	}
	return out
}

// Leaderboard returns results ordered newest-first for the /live dashboard.
func (s *Store) Leaderboard(limit int) []quiz.Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.results))
	for id := range s.results {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] > ids[j] })
	if limit > 0 && len(ids) > limit {
		ids = ids[:limit]
	}
	out := make([]quiz.Result, 0, len(ids))
	for _, id := range ids {
		out = append(out, s.results[id])
	}
	return out
}

func (s *Store) Events() ([]Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.f.Seek(0, 0); err != nil {
		return nil, err
	}
	defer s.f.Seek(0, 2)
	var out []Event
	sc := bufio.NewScanner(s.f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var e Event
		if err := json.Unmarshal(sc.Bytes(), &e); err == nil {
			out = append(out, e)
		}
	}
	return out, sc.Err()
}

func (s *Store) Close() error { return s.f.Close() }