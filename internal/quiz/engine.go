package quiz

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// Question is one multiple-choice item. Each option points at the modules it
// reveals an appetite for, so a quiz answers "which modules should we build your
// roadmap from", not "which of three personality types are you".
type Question struct {
	ID      int      `json:"id"`
	Text    string   `json:"text"`
	Options []Option `json:"options"`
}

type Option struct {
	Key     string   `json:"key"`
	Text    string   `json:"text"`
	Modules []string `json:"modules"`
}

type DataFile struct {
	Questions []Question `json:"questions"`
}

type Result struct {
	ID       string   `json:"id"`
	Answers  []string `json:"answers"`
	Counts   map[string]int `json:"counts"`
	Selected []string `json:"selected"`
}

func Load() (*DataFile, error) {
	var d DataFile
	if err := json.Unmarshal(Data, &d); err != nil {
		return nil, err
	}
	if len(d.Questions) == 0 {
		return nil, fmt.Errorf("quiz: no questions embedded")
	}
	for _, q := range d.Questions {
		if len(q.Options) < 2 {
			return nil, fmt.Errorf("quiz: question %d needs at least two options", q.ID)
		}
		for _, o := range q.Options {
			if len(o.Modules) == 0 {
				return nil, fmt.Errorf("quiz: question %d option %s maps to no module", q.ID, o.Key)
			}
		}
	}
	return &d, nil
}

func NewID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Score counts module mentions and picks the four strongest. Ties break on the
// module id so the same answers always produce the same roadmap.
func Score(answers []string, qs []Question, valid map[string]bool) (Result, error) {
	if len(answers) != len(qs) {
		return Result{}, fmt.Errorf("expected %d answers, got %d", len(qs), len(answers))
	}
	counts := map[string]int{}
	for i, a := range answers {
		q := qs[i]
		var opt *Option
		for j := range q.Options {
			if q.Options[j].Key == a {
				opt = &q.Options[j]
				break
			}
		}
		if opt == nil {
			return Result{}, fmt.Errorf("invalid answer %q for question %d", a, q.ID)
		}
		for _, m := range opt.Modules {
			if valid != nil && !valid[m] {
				return Result{}, fmt.Errorf("unknown module %q", m)
			}
			counts[m]++
		}
	}

	type pair struct {
		id string
		n  int
	}
	ranked := make([]pair, 0, len(counts))
	for id, n := range counts {
		ranked = append(ranked, pair{id, n})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].n != ranked[j].n {
			return ranked[i].n > ranked[j].n
		}
		return ranked[i].id < ranked[j].id
	})

	const want = 4
	selected := make([]string, 0, want)
	for _, r := range ranked {
		if len(selected) >= want {
			break
		}
		selected = append(selected, r.id)
	}
	// A quiz can only mention so many modules; pad deterministically from the
	// canonical order so every result yields a full roadmap.
	for _, id := range canonicalOrder {
		if len(selected) >= want {
			break
		}
		seen := false
		for _, s := range selected {
			if s == id {
				seen = true
				break
			}
		}
		if !seen {
			selected = append(selected, id)
		}
	}
	sort.Strings(selected)

	return Result{ID: NewID(), Answers: answers, Counts: counts, Selected: selected}, nil
}

// canonicalOrder is the fallback padding order (module id order).
var canonicalOrder = []string{"m1", "m2", "m3", "m4", "m5", "m6", "m7", "m8"}

// JoinSelection renders a selection as the comma form used in roadmap URLs.
func JoinSelection(sel []string) string {
	out := ""
	for i, s := range sel {
		if i > 0 {
			out += ","
		}
		out += s
	}
	return out
}