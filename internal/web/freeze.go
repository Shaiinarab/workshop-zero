package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"

	"workshop-zero/internal/plan"
	"workshop-zero/internal/quiz"
	"workshop-zero/internal/store"
	webfs "workshop-zero/web"
)

// Freeze exports a static snapshot: the same templates, with the quiz scored in
// the browser and roadmaps composed from the module library. No server, no keys,
// no network — this is what survives a dead tunnel or a blocked network.
func Freeze(out string) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "wz-events-*.jsonl")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	st, err := store.Open(tmpPath)
	if err != nil {
		return err
	}
	defer st.Close()

	// Seed one attempt so /result/{id} renders with real content.
	demo := demoResult()
	st.SaveResult(demo)

	lib, err := plan.LoadLibrary()
	if err != nil {
		return err
	}
	srv, err := New(st, lib)
	if err != nil {
		return err
	}
	h := srv.Routes()

	sel := lib.ParseSelection("")
	selPath := quiz.JoinSelection(sel)

	pages := []struct{ route, file string }{
		{"/", "index.html"},
		{"/program", "program/index.html"},
		{"/modules", "modules/index.html"},
		{"/proposal", "proposal/index.html"},
		{"/tools", "tools/index.html"},
		{"/faq", "faq/index.html"},
		{"/result/" + demo.ID, "result/index.html"},
		{"/result/" + demo.ID + "/og.svg", "result/og.svg"},
		{"/result/" + demo.ID + "/qr.png", "result/qr.png"},
		{"/roadmap/" + selPath + "?static=1", "roadmap/index.html"},
		{"/roadmap/" + selPath + "/export.md", "roadmap/roadmap.md"},
		{"/healthz", "healthz.json"},
	}

	srv.staticMode = true
	for _, p := range pages {
		srv.siteRoot = rootFor(p.file)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, p.route, nil)
		h.ServeHTTP(rec, req)
		if rec.Code >= 400 {
			return fmt.Errorf("freeze: %s → HTTP %d", p.route, rec.Code)
		}
		full := filepath.Join(out, p.file)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(full, rec.Body.Bytes(), 0o644); err != nil {
			return err
		}
	}

	// Every possible four-module selection gets a page, so any result the
	// browser-side scorer produces resolves to a real roadmap.
	for _, combo := range lib.Combinations(selectionSize) {
		p := "/" + quiz.JoinSelection(combo)
		file := filepath.Join("roadmap", p, "index.html")
		srv.siteRoot = rootFor(file)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/roadmap"+p+"?static=1", nil))
		if rec.Code >= 400 {
			return fmt.Errorf("freeze: roadmap %s → HTTP %d", p, rec.Code)
		}
		dir := filepath.Join(out, "roadmap", p)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "index.html"), rec.Body.Bytes(), 0o644); err != nil {
			return err
		}
	}

	// The frozen quiz is scored in the browser, so it renders from its own
	// template with the questions embedded as JSON.
	srv.siteRoot = "../"
	qr := httptest.NewRecorder()
	srv.r.Page(qr, "pages/quiz-static.html", map[string]any{
		"Title":         "آزمون مسیر",
		"Root":          srv.siteRoot,
		"Static":        true,
		"Total":         len(srv.quiz.Questions),
		"QuestionsJSON": template.JS(quizQuestionsJSON(srv.quiz.Questions)),
		"ModulesJSON":   template.JS(moduleListJSON(lib)),
	})
	if err := os.MkdirAll(filepath.Join(out, "quiz"), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "quiz", "index.html"), qr.Body.Bytes(), 0o644); err != nil {
		return err
	}

	// GitHub Pages runs Jekyll by default, which ignores underscore-prefixed paths
	// and can fail the build for a plain static export. This marker opts out.
	if err := os.WriteFile(filepath.Join(out, ".nojekyll"), nil, 0o644); err != nil {
		return err
	}

	return fs.WalkDir(webfs.FS, "static", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(p, "static/")
		if rel == "" || rel == p {
			if d.IsDir() {
				return os.MkdirAll(filepath.Join(out, "static"), 0o755)
			}
			return nil
		}
		dest := filepath.Join(out, "static", rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		f, err := webfs.FS.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dest, b, 0o644)
	})
}

func quizQuestionsJSON(qs []quiz.Question) string {
	b, err := json.Marshal(qs)
	if err != nil {
		return "[]"
	}
	return string(b)
}

const selectionSize = 4

// rootFor turns an output path into the relative prefix a page needs to reach
// the site root ("index.html" -> "./", "roadmap/a,b/index.html" -> "../../").
func rootFor(file string) string {
	d := path.Dir(filepath.ToSlash(file))
	if d == "." || d == "/" {
		return "./"
	}
	depth := len(strings.Split(d, "/"))
	return strings.Repeat("../", depth)
}

func moduleListJSON(lib *plan.Library) string {
	type lite struct {
		ID      string   `json:"id"`
		Title   string   `json:"title"`
		Tagline string   `json:"tagline"`
		Summary string   `json:"summary"`
		Tools   []string `json:"tools"`
	}
	out := make([]lite, 0, len(lib.All()))
	for _, m := range lib.All() {
		out = append(out, lite{m.ID, m.Title, m.Tagline, m.Summary, m.Tools})
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func demoResult() quiz.Result {
	return quiz.Result{
		ID:       "demo",
		Answers:  []string{"A", "A", "A", "A", "A", "A", "A", "A", "A", "A"},
		Counts:   map[string]int{"m1": 3, "m4": 3, "m2": 2, "m7": 2},
		Selected: []string{"m1", "m2", "m4", "m7"},
	}
}
