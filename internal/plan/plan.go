package plan

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed modules.json
var dataFS embed.FS

type Session struct {
	Title    string `json:"title"`
	Goal     string `json:"goal"`
	Artifact string `json:"artifact"`
}

type Module struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Tagline  string    `json:"tagline"`
	Summary  string    `json:"summary"`
	Concepts []string  `json:"concepts"`
	Tools    []string  `json:"tools"`
	Outcomes []string  `json:"outcomes"`
	Sessions []Session `json:"sessions"`
}

type Library struct {
	Modules []Module `json:"modules"`
}

// Item is one block of a roadmap.
type Item struct {
	N        int      `json:"n"`
	Kind     string   `json:"kind"` // workshop | core | module | studio | demo
	Title    string   `json:"title"`
	Goal     string   `json:"goal"`
	Artifact string   `json:"artifact"`
	ModuleID string   `json:"module_id,omitempty"`
	Tools    []string `json:"tools,omitempty"`
}

type Plan struct {
	Selection []string `json:"selection"`
	Modules   []Module `json:"modules"`
	Items     []Item   `json:"items"`
	Track     string   `json:"track"`
}

// selectionSize is how many modules a roadmap is built from.
const selectionSize = 4

func LoadLibrary() (*Library, error) {
	b, err := dataFS.ReadFile("modules.json")
	if err != nil {
		return nil, err
	}
	var lib Library
	if err := json.Unmarshal(b, &lib); err != nil {
		return nil, err
	}
	if len(lib.Modules) == 0 {
		return nil, fmt.Errorf("plan: no modules embedded")
	}
	for _, m := range lib.Modules {
		if len(m.Sessions) == 0 {
			return nil, fmt.Errorf("plan: module %s has no sessions", m.ID)
		}
	}
	return &lib, nil
}

// All returns every module in canonical order.
func (l *Library) All() []Module { return l.Modules }

func (l *Library) Get(id string) (Module, bool) {
	for _, m := range l.Modules {
		if m.ID == id {
			return m, true
		}
	}
	return Module{}, false
}

func (l *Library) IDs() []string {
	out := make([]string, 0, len(l.Modules))
	for _, m := range l.Modules {
		out = append(out, m.ID)
	}
	return out
}

// ParseSelection turns "m1,m3,m5" into a validated, canonical-ordered selection.
// Unknown ids are dropped; missing slots are padded from the canonical order, so
// a short or dirty URL still yields a complete roadmap.
func (l *Library) ParseSelection(raw string) []string {
	seen := map[string]bool{}
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == '-' || r == ' ' }) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := l.Get(part); ok {
			seen[part] = true
		}
	}
	// Canonical order always, so a selection has exactly one representation.
	out := make([]string, 0, selectionSize)
	for _, id := range l.IDs() {
		if seen[id] {
			out = append(out, id)
		}
	}
	for _, id := range l.IDs() {
		if len(out) >= selectionSize {
			break
		}
		if !seen[id] {
			out = append(out, id)
		}
	}
	if len(out) > selectionSize {
		out = out[:selectionSize]
	}
	return out
}

// Compose builds the single 15-block roadmap: one workshop, a shared core, the
// chosen modules, a project studio, and Demo Day. Only the middle eight blocks
// change with the selection — the shape is always the same.
func (l *Library) Compose(selection []string) Plan {
	sel := l.canonical(selection)
	mods := make([]Module, 0, len(sel))
	for _, id := range sel {
		if m, ok := l.Get(id); ok {
			mods = append(mods, m)
		}
	}

	var items []Item
	n := 0
	add := func(kind, title, goal, artifact, moduleID string, tools []string) {
		items = append(items, Item{N: n, Kind: kind, Title: title, Goal: goal, Artifact: artifact, ModuleID: moduleID, Tools: tools})
		n++
	}

	add("workshop", "کارگاه صفر: مسیرتو پیدا کن",
		"سه ایستگاه، آزمون مسیر و ساختن نقشه راه شخصی", "نتیجه آزمون + نقشه راه", "", nil)

	add("core", "زمینه مشترک: نقشه راه شخصی",
		"تعیین هدف، انتخاب ابزار و ساختن برنامه یادگیری", "برنامه شخصی", "", nil)
	add("core", "زمینه مشترک: اعتبارسنجی و اتکای درست",
		"بررسی خروجی، کشف خطا و نوشتن قواعد استفاده", "چک‌لیست بررسی", "", nil)

	for _, m := range mods {
		for _, s := range m.Sessions {
			add("module", s.Title, s.Goal, s.Artifact, m.ID, m.Tools)
		}
	}

	add("studio", "استودیوی پروژه ۱: تعریف مسئله",
		"انتخاب مسئله، ابزار و معیار موفقیت", "پروپوزال پروژه", "", nil)
	add("studio", "استودیوی پروژه ۲: ساخت و تکرار",
		"ساخت نسخه اول و بهبود بر اساس بازخورد", "نمونه کارا", "", nil)
	add("studio", "استودیوی پروژه ۳: بازخورد و آماده‌سازی",
		"دریافت بازخورد، آماده‌سازی ارائه", "نسخه نهایی", "", nil)

	add("demo", "Demo Day", "ارائه دستاوردها و جمع‌بندی", "ارائه نهایی", "", nil)

	return Plan{Selection: sel, Modules: mods, Items: items, Track: strings.Join(sel, "+")}
}

func (l *Library) canonical(sel []string) []string {
	seen := map[string]bool{}
	for _, s := range sel {
		seen[s] = true
	}
	out := make([]string, 0, len(sel))
	for _, id := range l.IDs() {
		if seen[id] {
			out = append(out, id)
		}
	}
	return out
}

// Headline is the one-line description of a plan, used in the hero and exports.
func (p Plan) Headline() string {
	titles := make([]string, 0, len(p.Modules))
	for _, m := range p.Modules {
		titles = append(titles, m.Title)
	}
	return "نقشه راه: " + strings.Join(titles, " · ")
}

func (p Plan) Count() int { return len(p.Items) }

// Markdown exports the plan as a shareable document.
func (p Plan) Markdown() string {
	var b strings.Builder
	b.WriteString("# " + p.Headline() + "\n\n")
	b.WriteString("کارگاه صفر — موسسه آموزش عالی زند شیراز\n\n")
	for _, it := range p.Items {
		fmt.Fprintf(&b, "## %s — %s\n\n", faNum(it.N), it.Title)
		fmt.Fprintf(&b, "- هدف: %s\n", it.Goal)
		fmt.Fprintf(&b, "- خروجی: %s\n", it.Artifact)
		if len(it.Tools) > 0 {
			fmt.Fprintf(&b, "- ابزارها: %s\n", strings.Join(it.Tools, " | "))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// Marp exports the plan as a Marp slide deck.
func (p Plan) Marp() string {
	var b strings.Builder
	b.WriteString("---\nmclass: lead\n---\n\n# " + p.Headline() + "\n\nکارگاه صفر\n\n---\n\n")
	for _, it := range p.Items {
		fmt.Fprintf(&b, "## %s — %s\n\n%s\n\nخروجی: %s\n\n---\n\n", faNum(it.N), it.Title, it.Goal, it.Artifact)
	}
	return b.String()
}

func faNum(n int) string {
	d := []rune("۰۱۲۳۴۵۶۷۸۹")
	var out strings.Builder
	for _, r := range fmt.Sprintf("%d", n) {
		if r >= '0' && r <= '9' {
			out.WriteRune(d[r-'0'])
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}

// SortModules is a small helper used by the library page to order modules.
func SortModules(ms []Module) {
	sort.Slice(ms, func(i, j int) bool { return ms[i].ID < ms[j].ID })
}
// Combinations returns every k-subset of the module ids in canonical order.
// The frozen build uses it to emit a roadmap page per possible selection.
func (l *Library) Combinations(k int) [][]string {
	ids := l.IDs()
	var out [][]string
	var walk func(start int, cur []string)
	walk = func(start int, cur []string) {
		if len(cur) == k {
			out = append(out, append([]string(nil), cur...))
			return
		}
		for i := start; i < len(ids); i++ {
			walk(i+1, append(cur, ids[i]))
		}
	}
	walk(0, nil)
	return out
}
