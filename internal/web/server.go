package web

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	qrcode "github.com/skip2/go-qrcode"

	"workshop-zero/content"
	"workshop-zero/internal/plan"
	"workshop-zero/internal/quiz"
	"workshop-zero/internal/store"
	webfs "workshop-zero/web"
)

type Server struct {
	store *store.Store
	lib   *plan.Library
	hub   *Hub
	r     *Renderer
	quiz  *quiz.DataFile
	base  string

	mu      sync.Mutex
	streams map[string]*streamState

	// Set by Freeze so the same templates emit subpath-relative URLs.
	siteRoot   string
	staticMode bool
}

type streamState struct {
	mu   sync.Mutex
	plan plan.Plan
	sent int
	done bool
}

func New(st *store.Store, lib *plan.Library) (*Server, error) {
	r, err := NewRenderer()
	if err != nil {
		return nil, err
	}
	qd, err := quiz.Load()
	if err != nil {
		return nil, err
	}
	return &Server{
		store:   st,
		lib:     lib,
		hub:     NewHub(),
		r:       r,
		quiz:    qd,
		base:    envOr("BASE_URL", "http://localhost:8080"),
		streams: map[string]*streamState{},
	}, nil
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	staticFS, _ := fs.Sub(webfs.FS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	mux.HandleFunc("GET /{$}", s.handleCover)
	mux.HandleFunc("GET /healthz", s.handleHealth)

	mux.HandleFunc("GET /program", s.handleProgram)
	mux.HandleFunc("GET /workshop", s.handleWorkshop)
	mux.HandleFunc("GET /modules", s.handleModules)
	mux.HandleFunc("GET /proposal", s.handleProposal)
	mux.HandleFunc("GET /tools", s.handleTools)
	mux.HandleFunc("GET /faq", s.handleFAQ)

	mux.HandleFunc("GET /quiz", s.handleQuizStart)
	mux.HandleFunc("GET /quiz/{id}", s.handleQuizPage)
	mux.HandleFunc("GET /quiz/{id}/q/{n}", s.handleQuizQuestion)
	mux.HandleFunc("POST /quiz/{id}/answer", s.handleQuizAnswer)
	mux.HandleFunc("POST /quiz/{id}/reset", s.handleQuizReset)

	mux.HandleFunc("GET /result/{id}", s.handleResult)
	mux.HandleFunc("GET /result/{id}/qr.png", s.handleQR)
	mux.HandleFunc("GET /result/{id}/og.svg", s.handleOG)

	mux.HandleFunc("GET /roadmap", s.handleRoadmapDefault)
	mux.HandleFunc("GET /roadmap/{sel}", s.handleRoadmap)
	mux.HandleFunc("GET /roadmap/{sel}/session/{n}", s.handleSessionFragment)
	mux.HandleFunc("POST /roadmap/{sel}/done", s.handleDone)
	mux.HandleFunc("GET /roadmap/{sel}/export.md", s.handleExportMD)
	mux.HandleFunc("GET /roadmap/{sel}/export.marp.md", s.handleExportMarp)

	mux.HandleFunc("GET /live", s.handleLive)
	mux.HandleFunc("GET /live/export.jsonl", s.handleLiveExport)

	mux.HandleFunc("GET /events", s.handleEvents)
	mux.HandleFunc("POST /events/click", s.handleClick)

	return mux
}

// ------------------------------------------------------------------ helpers

func isHTMX(r *http.Request) bool { return r.Header.Get("HX-Request") == "true" }

func envOr(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i != -1 {
		host = host[:i]
	}
	return host
}

func (s *Server) page(w http.ResponseWriter, name, title string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	data["Title"] = title
	data["Root"] = s.siteRoot
	data["Static"] = s.staticMode
	s.r.Page(w, name, data)
}

func (s *Server) notFound(w http.ResponseWriter, r *http.Request) {
	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(200)
		return
	}
	http.Error(w, "not found", http.StatusNotFound)
}

// ------------------------------------------------------------------ cover

func (s *Server) handleCover(w http.ResponseWriter, r *http.Request) {
	_ = s.store.Append(store.Event{Kind: "page_view", Path: "/", IP: clientIP(r)})
	lib := s.lib.All()
	picked := []plan.Module{}
	for i := 0; i < 4 && i < len(lib); i++ {
		picked = append(picked, lib[i])
	}
	s.page(w, "pages/cover.html", "کارگاه صفر", map[string]any{
		"Modules":  len(lib),
		"Sessions": len(s.lib.Compose(s.lib.ParseSelection("")).Items),
		"Stations": markdownDir("stations"),
		"Proposal": markdownFile("proposal/proposal.md"),
		"Preview":  picked,
		"Finished": len(s.store.Leaderboard(0)),
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "events": s.store.Count()})
}

// ------------------------------------------------------------------ program

func (s *Server) handleProgram(w http.ResponseWriter, r *http.Request) {
	s.page(w, "pages/program.html", "برنامه", map[string]any{
		"Stations": markdownDir("stations"),
		"Agenda":   agendaState(),
	})
}

func (s *Server) handleWorkshop(w http.ResponseWriter, r *http.Request) {
	s.page(w, "pages/workshop.html", "راهنمای کارگاه صفر", map[string]any{
		"Guide": markdownFile("workshops/general-zero-workshop.md"),
	})
}

func (s *Server) handleModules(w http.ResponseWriter, r *http.Request) {
	s.page(w, "pages/modules.html", "ماژول‌ها", map[string]any{
		"Modules": s.lib.All(),
	})
}

func (s *Server) handleFAQ(w http.ResponseWriter, r *http.Request) {
	s.page(w, "pages/faq.html", "پرسش‌های متداول", nil)
}

// handleProposal renders the teacher-facing program proposal (content/proposal).
func (s *Server) handleProposal(w http.ResponseWriter, r *http.Request) {
	s.page(w, "pages/proposal.html", "پیشنهاد برنامه", map[string]any{
		"Body": markdownFile("proposal/proposal.md"),
	})
}

// ------------------------------------------------------------------ quiz

func (s *Server) handleQuizStart(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/quiz/"+quiz.NewID(), http.StatusFound)
}

func (s *Server) handleQuizPage(w http.ResponseWriter, r *http.Request) {
	id := s.mustID(w, r)
	if id == "" {
		return
	}
	total := len(s.quiz.Questions)
	s.store.Reset(id)
	_ = s.store.Append(store.Event{Kind: "quiz_start", Path: "/quiz/" + id, IP: clientIP(r)})
	s.page(w, "pages/quiz.html", "آزمون مسیر", map[string]any{
		"ID": id, "Total": total, "Answered": 0, "Pct": 0,
	})
}

func (s *Server) mustID(w http.ResponseWriter, r *http.Request) string {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" || len(id) > 40 || strings.ContainsAny(id, "/\\ ") {
		http.Error(w, "bad id", http.StatusBadRequest)
		return ""
	}
	return id
}

func (s *Server) handleQuizQuestion(w http.ResponseWriter, r *http.Request) {
	id := s.mustID(w, r)
	if id == "" {
		return
	}
	n, err := strconv.Atoi(r.PathValue("n"))
	if err != nil || n < 1 || n > len(s.quiz.Questions) {
		s.notFound(w, r)
		return
	}
	total := len(s.quiz.Questions)
	answered := len(s.store.Answers(id))
	if answered > total {
		answered = total
	}
	pct := answered * 100 / total
	data := map[string]any{
		"ID": id, "N": n, "Total": total, "Q": s.quiz.Questions[n-1],
		"Pct": pct, "Answered": answered,
	}
	if isHTMX(r) {
		data["Root"] = s.siteRoot
		data["Static"] = s.staticMode
		s.r.Fragment(w, "fragments/quiz-question.html", data)
		return
	}
	s.page(w, "pages/quiz-question.html", fmt.Sprintf("سؤال %d از %d", n, total), data)
}

func (s *Server) handleQuizReset(w http.ResponseWriter, r *http.Request) {
	id := s.mustID(w, r)
	if id == "" {
		return
	}
	s.store.Reset(id)
	if isHTMX(r) {
		w.Header().Set("HX-Redirect", "/quiz/"+id)
		w.WriteHeader(200)
		return
	}
	http.Redirect(w, r, "/quiz/"+id, http.StatusSeeOther)
}

func (s *Server) handleQuizAnswer(w http.ResponseWriter, r *http.Request) {
	id := s.mustID(w, r)
	if id == "" {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	n, _ := strconv.Atoi(r.FormValue("n"))
	answer := strings.TrimSpace(r.FormValue("answer"))
	if n < 1 || n > len(s.quiz.Questions) || answer == "" {
		http.Error(w, "bad answer", http.StatusUnprocessableEntity)
		return
	}
	if !validOption(s.quiz.Questions[n-1], answer) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		s.r.Fragment(w, "fragments/quiz-error.html", map[string]any{"N": n, "Total": len(s.quiz.Questions)})
		return
	}

	answers := append(s.store.Answers(id), answer)
	s.store.Answer(id, n, answer)

	total := len(s.quiz.Questions)
	pct := len(answers) * 100 / total
	if pct > 100 {
		pct = 100
	}

	if n >= total {
		valid := map[string]bool{}
		for _, id := range s.lib.IDs() {
			valid[id] = true
		}
		res, err := quiz.Score(answers, s.quiz.Questions, valid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res.ID = id
		s.store.SaveResult(res)
		s.hub.Send("", s.counterFragment())

		if isHTMX(r) {
			w.Header().Set("HX-Redirect", "/result/"+id)
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Redirect(w, r, "/result/"+id, http.StatusSeeOther)
		return
	}

	next := n + 1
	data := map[string]any{
		"ID": id, "N": next, "Total": total, "Q": s.quiz.Questions[next-1],
		"Pct": pct, "Answered": len(answers),
	}
	if isHTMX(r) {
		var b strings.Builder
		frag, err := s.r.RenderString("fragments/quiz-question.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		prog, err := s.r.RenderString("fragments/quiz-progress.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		b.WriteString(frag)
		fmt.Fprintf(&b, `<template hx type="partial" hx-target="#quiz-progress" hx-swap="innerHTML">%s</template>`, prog)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(b.String()))
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/quiz/%s/q/%d", id, next), http.StatusSeeOther)
}

func validOption(q quiz.Question, key string) bool {
	for _, o := range q.Options {
		if o.Key == key {
			return true
		}
	}
	return false
}

// ------------------------------------------------------------------ result

func (s *Server) handleResult(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	res, ok := s.store.Result(id)
	if !ok {
		s.notFound(w, r)
		return
	}
	mods := make([]plan.Module, 0, len(res.Selected))
	for _, mid := range res.Selected {
		if m, ok := s.lib.Get(mid); ok {
			mods = append(mods, m)
		}
	}
	_ = s.store.Append(store.Event{Kind: "result_view", Path: "/result/" + id, IP: clientIP(r), Meta: map[string]any{"sel": res.Selected}})

	sel := quiz.JoinSelection(res.Selected)
	s.page(w, "pages/result.html", "نتیجه", map[string]any{
		"ID": id, "R": res, "Modules": mods, "Sel": sel,
		"ShareURL": s.base + "/result/" + id,
	})
}

func (s *Server) handleQR(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := s.store.Result(id); !ok {
		http.NotFound(w, r)
		return
	}
	png, err := qrcode.Encode(s.base+"/result/"+id, qrcode.Medium, 280)
	if err != nil {
		http.Error(w, "qr unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(png)
}

func (s *Server) handleOG(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	res, ok := s.store.Result(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	mods := make([]plan.Module, 0, len(res.Selected))
	for _, mid := range res.Selected {
		if m, ok := s.lib.Get(mid); ok {
			mods = append(mods, m)
		}
	}
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="630" viewBox="0 0 1200 630">`)
	b.WriteString(`<rect width="1200" height="630" fill="#141018"/>`)
	b.WriteString(`<rect x="0" y="0" width="1200" height="12" fill="#C8452D"/>`)
	fmt.Fprintf(&b, `<text x="80" y="140" fill="#9A9370" font-family="sans-serif" font-size="32" direction="rtl">کارگاه صفر — موسسه آموزش عالی زند</text>`)
	fmt.Fprintf(&b, `<text x="80" y="250" fill="#FBF8F3" font-family="sans-serif" font-size="72" font-weight="bold" direction="rtl">نقشه راه شخصی</text>`)
	y := 340
	for _, m := range mods {
		fmt.Fprintf(&b, `<text x="80" y="%d" fill="#E8DFC8" font-family="sans-serif" font-size="38" direction="rtl">• %s</text>`, y, xmlEscape(m.Title))
		y += 62
	}
	b.WriteString(`</svg>`)

	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write([]byte(b.String()))
}

func xmlEscape(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

// ------------------------------------------------------------------ roadmap

// handleRoadmapDefault gives the bare /roadmap URL a real plan instead of a 404.
func (s *Server) handleRoadmapDefault(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/roadmap/"+quiz.JoinSelection(s.lib.ParseSelection("")), http.StatusFound)
}

func (s *Server) handleRoadmap(w http.ResponseWriter, r *http.Request) {
	sel := s.lib.ParseSelection(r.PathValue("sel"))
	p := s.lib.Compose(sel)
	key := "roadmap:" + strings.Join(sel, "-")

	s.mu.Lock()
	st, ok := s.streams[key]
	if !ok {
		st = &streamState{plan: p}
		s.streams[key] = st
	}
	s.mu.Unlock()

	static := r.URL.Query().Get("static") == "1"
	_ = s.store.Append(store.Event{Kind: "roadmap_view", Path: "/roadmap/" + r.PathValue("sel"), IP: clientIP(r), Meta: map[string]any{"sel": sel}})

	s.page(w, "pages/roadmap.html", "نقشه راه", map[string]any{
		"Sel":     quiz.JoinSelection(sel),
		"Stream":  key,
		"Plan":    p,
		"Static":  static,
		"Done":    s.store.Done(key),
		"DoneCnt": countDone(s.store.Done(key)),
		"Total":   len(p.Items),
	})
}

func countDone(m map[int]bool) int {
	n := 0
	for _, v := range m {
		if v {
			n++
		}
	}
	return n
}

func (s *Server) handleSessionFragment(w http.ResponseWriter, r *http.Request) {
	sel := s.lib.ParseSelection(r.PathValue("sel"))
	p := s.lib.Compose(sel)
	n, err := strconv.Atoi(r.PathValue("n"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var item *plan.Item
	for i := range p.Items {
		if p.Items[i].N == n {
			item = &p.Items[i]
			break
		}
	}
	if item == nil {
		s.notFound(w, r)
		return
	}
	key := "roadmap:" + strings.Join(sel, "-")
	s.r.Fragment(w, "fragments/session-detail.html", map[string]any{
		"Sel": quiz.JoinSelection(sel), "I": item, "Done": s.store.Done(key)[item.N],
	})
}

func (s *Server) handleDone(w http.ResponseWriter, r *http.Request) {
	sel := s.lib.ParseSelection(r.PathValue("sel"))
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	n, _ := strconv.Atoi(r.FormValue("n"))
	key := "roadmap:" + strings.Join(sel, "-")
	done := s.store.Done(key)
	done[n] = !done[n]
	s.store.SetDone(key, n, done[n])

	p := s.lib.Compose(sel)
	s.r.Fragment(w, "fragments/roadmap-progress.html", map[string]any{
		"Done": done, "DoneCnt": countDone(done), "Total": len(p.Items),
	})
}

func (s *Server) handleExportMD(w http.ResponseWriter, r *http.Request) {
	p := s.lib.Compose(s.lib.ParseSelection(r.PathValue("sel")))
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="roadmap.md"`)
	_, _ = w.Write([]byte(p.Markdown()))
}

func (s *Server) handleExportMarp(w http.ResponseWriter, r *http.Request) {
	p := s.lib.Compose(s.lib.ParseSelection(r.PathValue("sel")))
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="roadmap-marp.md"`)
	_, _ = w.Write([]byte(p.Marp()))
}

// ------------------------------------------------------------------ live

func (s *Server) liveGate(w http.ResponseWriter, r *http.Request) bool {
	token := envOr("LIVE_TOKEN", "")
	if token == "" {
		return true
	}
	if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("t")), []byte(token)) != 1 {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func (s *Server) handleLiveExport(w http.ResponseWriter, r *http.Request) {
	if !s.liveGate(w, r) {
		return
	}
	events, err := s.store.Events()
	if err != nil {
		http.Error(w, "export failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-ndjson; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="events.jsonl"`)
	enc := json.NewEncoder(w)
	for _, e := range events {
		if err := enc.Encode(e); err != nil {
			return
		}
	}
}

func (s *Server) handleLive(w http.ResponseWriter, r *http.Request) {
	if !s.liveGate(w, r) {
		return
	}
	board := s.store.Leaderboard(10)
	dist := s.store.Distribution()
	max := 0
	for _, v := range dist {
		if v > max {
			max = v
		}
	}
	ids := make([]string, 0, len(dist))
	for id := range dist {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	_ = s.store.Append(store.Event{Kind: "live_view", IP: clientIP(r)})

	s.page(w, "pages/live.html", "داشبورد مربی", map[string]any{
		"Total":   s.store.Count(),
		"Board":   board,
		"Dist":    dist,
		"DistMax": max,
		"IDs":     ids,
		"Lib":     s.lib,
		"Agenda":  agendaState(),
	})
}

// agendaState drives the 180-minute rotation. It is derived from the wall clock
// so every screen agrees without a server-side timer.
func agendaState() map[string]any {
	type item struct {
		Start int
		End   int
		Title string
		Note  string
	}
	items := []item{
		{0, 15, "قلاب داستانی و مسئله واقعی", "نمایش چند خروجی متفاوت و طرح پرسش آغازین"},
		{15, 35, "نقشه جهان هوش مصنوعی", "شناخت AI، مدل، داده، ابزار، قابلیت و API"},
		{35, 55, "آزمایش زنده", "مقایسه دستور مبهم با درخواست زمینه‌دار و معیارمند"},
		{55, 110, "آزمایشگاه هدایت‌شده", "تمرین فردی یا دونفره با چرخه برنامه‌ریزی، ساخت، بررسی و تکرار"},
		{110, 120, "استراحت و گالری", "نمایش سریع خروجی‌ها و مرور تجربه‌ها"},
		{120, 155, "چالش تیمی", "تبدیل یک مسئله انتخابی به نمونه اولیه"},
		{155, 170, "نمایش و نقد", "ارائه کوتاه، بازخورد همتا و راستی‌آزمایی"},
		{170, 180, "برگه خروج و مسیر بعدی", "یادآوری آموخته‌ها و انتخاب قدم بعدی"},
	}
	return map[string]any{"Total": 180, "Items": items}
}

// ------------------------------------------------------------------ tools

func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	s.page(w, "pages/tools.html", "ابزارها", map[string]any{
		"Tools": markdownFile("resources/tools.md"),
		"Repos": markdownFile("resources/repos.md"),
	})
}

// ------------------------------------------------------------------ fragments

func (s *Server) counterFragment() string {
	html, err := s.r.RenderString("fragments/counter.html", map[string]any{"N": len(s.store.Leaderboard(0))})
	if err != nil {
		return ""
	}
	return html
}

func (s *Server) handleClick(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = s.store.Append(store.Event{Kind: "click", Path: r.FormValue("path"), IP: clientIP(r)})
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("stream")

	s.mu.Lock()
	st := s.streams[id]
	s.mu.Unlock()

	if st != nil && strings.HasPrefix(id, "roadmap:") {
		s.pump(w, r, id, st)
		return
	}
	s.hub.Serve(w, r, id, s.counterFragment())
}

// pump streams one roadmap block at a time. Progress lives on the stream state,
// so a reconnect resumes instead of restarting.
func (s *Server) pump(w http.ResponseWriter, r *http.Request, id string, st *streamState) {
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

	for {
		select {
		case <-r.Context().Done():
			return
		default:
		}

		st.mu.Lock()
		if st.sent < len(st.plan.Items) {
			it := st.plan.Items[st.sent]
			st.sent++
			st.mu.Unlock()

			html, err := s.r.RenderString("fragments/plan-item.html", map[string]any{"I": it, "Sel": st.plan.Track})
			if err == nil {
				fmt.Fprint(w, "data: "+oneLine(html)+"\n\n")
				flusher.Flush()
			}
			select {
			case <-r.Context().Done():
				return
			case <-time.After(120 * time.Millisecond):
			}
			continue
		}
		if !st.done {
			st.done = true
			st.mu.Unlock()
			fmt.Fprint(w, "event: plan-done\ndata: done\n\n")
			flusher.Flush()
			continue
		}
		st.mu.Unlock()

		select {
		case <-r.Context().Done():
			return
		case <-time.After(25 * time.Second):
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

// ------------------------------------------------------------------ content

func markdownDir(dir string) []map[string]string {
	entries, err := fs.ReadDir(content.FS, dir)
	if err != nil {
		return nil
	}
	out := []map[string]string{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == "README.md" {
			continue
		}
		b, err := content.FS.ReadFile(dir + "/" + e.Name())
		if err != nil {
			continue
		}
		out = append(out, map[string]string{"Name": strings.TrimSuffix(e.Name(), ".md"), "Body": string(b)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["Name"] < out[j]["Name"] })
	return out
}

func markdownFile(name string) string {
	b, err := content.FS.ReadFile(name)
	if err != nil {
		return ""
	}
	return string(b)
}

var _ = context.Background
