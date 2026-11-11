package web

import (
	"bytes"
	"workshop-zero/internal/plan"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	webfs "workshop-zero/web"
)

// Renderer holds every template parsed once at boot. All HTML is produced here
// or in a template — never from model output — which is the injection boundary.
type Renderer struct {
	tmpl *template.Template
}

func NewRenderer() (*Renderer, error) {
	funcs := template.FuncMap{
		"faNum": faNum,
		"pct":   func(score, total int) int { if total == 0 { return 0 }; return score * 100 / total },
		"sub":   func(a, b int) int { return a - b },
		"add":   func(a, b int) int { return a + b },
		"eq":    func(a, b any) bool { return fmt.Sprint(a) == fmt.Sprint(b) },
		"seq":   func(n int) []int { out := make([]int, n); for i := range out { out[i] = i + 1 }; return out },
		"title": trackTitleFA,
		"short": trackShortFA,
		"badge": layerClass,
		"join":  func(parts []string, sep string) string { return strings.Join(parts, sep) },
		"md":    mdLite,
		"item": func(sel string, it plan.Item) map[string]any {
			return map[string]any{"Sel": sel, "I": it}
		},
	}
	t := template.New("").Funcs(funcs)
	err := fs.WalkDir(webfs.FS, "templates", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".html") {
			return err
		}
		b, err := webfs.FS.ReadFile(p)
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(p, "templates/")
		_, err = t.New(name).Parse(string(b))
		return err
	})
	if err != nil {
		return nil, err
	}
	return &Renderer{tmpl: t}, nil
}

// Page renders a full page (used for no-JS requests and full navigations).
func (r *Renderer) Page(w http.ResponseWriter, name string, data any) {
	var buf bytes.Buffer
	if err := r.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(buf.Bytes())
}

// Fragment renders a swap-target fragment.
func (r *Renderer) Fragment(w http.ResponseWriter, name string, data any) {
	var buf bytes.Buffer
	if err := r.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(buf.Bytes())
}

// RenderString is used where a fragment must travel over SSE instead of a
// response body.
func (r *Renderer) RenderString(name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := r.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
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

func layerClass(layer string) string {
	switch layer {
	case "✅":
		return "layer-free"
	case "🔵":
		return "layer-optional"
	case "🟡":
		return "layer-vpn"
	}
	return ""
}

func trackTitleFA(t string) string {
	switch t {
	case "builder":
		return "سازنده (Builder)"
	case "thinker":
		return "متفکر (Thinker)"
	case "maker":
		return "خلاق (Maker)"
	}
	return t
}

func trackShortFA(t string) string {
	switch t {
	case "builder":
		return "سازنده"
	case "thinker":
		return "متفکر"
	case "maker":
		return "خلاق"
	}
	return t
}