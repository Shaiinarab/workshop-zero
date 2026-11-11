# PROMPT.md — Workshop Zero Preview (v3)
# «کارگاه صفر: مسیرتو پیدا کن» — one-binary Go + htmx 4 + TypeScript 7
Verified against the live web on 2026-11-11. Re-verify anything marked [VERIFY].

## 0) Role & Mission
You are a senior Go engineer + htmx craftsman + content architect. Build a
**single-binary web app** that previews the university AI Workshops program:
- Scope = **Workshop 00 only** (the 3-hour orientation from PROMPT-v2 §4)
  **plus a dynamic roadmap engine** for everything that comes after (3 tracks,
  14 sessions, personalized per quiz result, optionally AI-generated).
- All UI text **Persian, RTL** (`lang="fa" dir="rtl"`), tech terms in English,
  Vazirmatn font **vendored locally** (no CDN at runtime — must load in Iran
  without VPN). Filenames/code/commits in English.
- Philosophy: HTML over the wire. The server renders fragments; the browser
  stays dumb and fast. No SPA, no React, no client-side router.
- Every demo must work with **zero live VPN dependency** and **zero paid API**.

## 1) Verified facts (2026-11-11) — do not code from old memory
| Fact | Detail | Source |
|---|---|---|
| htmx 4.0.0 STABLE | Released 2026-08-28. npm: `htmx.org@4.0.0` — **`latest` tag is still 2.x until early 2027**, so pin the exact version. | https://four.htmx.org/announcements/2026-08-28-htmx-4.0.0-is-released |
| htmx 4 breaking | (a) inheritance is opt-in via `:inherited` suffix (`hx-target:inherited`); (b) ALL responses swap incl. 4xx/5xx (except 204/304) — use `hx-status:422` for form errors; (c) events renamed `htmx:phase:action` (`htmx:after:swap`, `htmx:config:request`); errors → `htmx:error` / `htmx:response:error`; (d) no localStorage history cache; (e) `hx-disable`→`hx-ignore`, old `hx-disabled-elt`→`hx-disable`; (f) `hx-vars`/`hx-prompt`/`hx-ext` removed; extensions are plain `<script>` includes; (g) fetch() internals, 60s default timeout, queueing only via `hx-sync`. | https://four.htmx.org/docs/whats-new-in-htmx-4 |
| htmx 4 new toys | `innerMorph`/`outerMorph` swaps (idiomorph built-in); `<hx-partial>` multi-target responses; `hx-action`+`hx-method` (progressive enhancement w/ native form fallback); `hx-live` (DOM-reactive micro-scripting, replaces Alpine for our scale); rewritten `hx-sse` / `hx-ws` + new `hx-multipart` streaming; `hx-preload` (hover); `hx-download`; `htmax.js` bundle. | https://four.htmx.org/docs |
| htmx ships LLM skills | Official skill files: `htmx-guidance`, `htmx-debugging`, `htmx-extension-authoring`, `htmx-upgrade-from-htmx2`. **Fetch and read them before coding.** Upgrade checker: `npx htmx.org@4.0.0 upgrade-check -- ./web/templates` | same announcement |
| TypeScript 7 STABLE | GA 2026-07-08 as `typescript@7` with normal `tsc` (Go-native port, ~10× faster, shared-memory parallel). 7.1 Beta 2026-10-20 restores compiler API (we don't need it — compiler-only use). `tsgo` name lives on only in `@typescript/native-preview` nightlies. | https://devblogs.microsoft.com/typescript/announcing-typescript-7-0/ |
| TS7 defaults | `strict:true`, `module:esnext`, `rootDir:./` (set `rootDir:"src"`), `types:[]` (list explicitly), es5/node10 resolution removed, `--checkers`/`--builders`/`--singleThreaded` flags. No bundler needed: emit plain ES modules. | https://devblogs.microsoft.com/typescript/announcing-typescript-7-0-beta/ |
| Go 1.27 | Released Aug 2026: generic methods, `encoding/json/v2`, stdlib `uuid`, faster allocation. Use stdlib `net/http` ServeMux patterns (`mux.Handle("POST /quiz/{id}/answer", …)`), `embed.FS`, `html/template`, `http.ResponseController` for SSE flushing. **No web framework.** | https://go.dev/doc/go1.27 |
| Cloudflare Quick Tunnel | `cloudflared tunnel --url http://localhost:8080` → public `*.trycloudflare.com` URL. **Free, no account, no domain.** Ephemeral — perfect for class demos. | https://developers.cloudflare.com/tunnel/get-started/quick-tunnels/ |
| Render free tier | **No credit card.** 750 instance-hours/month, spins down after 15 min idle (cold start ~1 min — keep binary tiny). Native Go or Docker, auto-deploy from GitHub. | https://render.com/docs/free |
| Hosting traps 2026 | Koyeb free now **requires a payment card**; HF Spaces Docker/Gradio **paid since July 2026**; Fly ≈ $2/mo. Do NOT use them. GitHub Pages is free and Iran-accessible (GitHub OFAC license since 2021). | https://www.srvrlss.io/provider/koyeb/ , https://snapdeploy.dev/blog/free-cloud-deployment-platforms-2026-comparison |
| Free LLM APIs (no KYC) | From PROMPT-v2: **OpenCode Zen** (free models, OpenAI-compatible), **NVIDIA NIM** (~5000 credits, 40 RPM, no card), **Zhipu GLM-4.7-Flash** (free, OpenAI SDK compatible). OpenRouter = backup only (50 req/day free). Local fallback: Ollama. | PROMPT-v2 §1 + https://build.nvidia.com , https://open.bigmodel.cn |

## 2) Agent learning tasks (BEFORE writing code)
1. Fetch & study: `https://four.htmx.org/docs/whats-new-in-htmx-4`, the `hx-sse`
   and `hx-live` extension docs, and the official htmx LLM skill files. [VERIFY
   exact `hx-sse` attribute syntax in v4 — it is namespaced, e.g. `hx-sse:connect`;
   confirm from docs, not from htmx 2 memory.]
2. Read the TS 7.0 announcement + your project's `tsconfig` against §1 defaults.
3. Read `https://go.dev/doc/go1.27` (json/v2, generic methods) — use them.
4. Vendor assets at build time (never at runtime): `htmx.org@4.0.0` dist
   (`htmx.min.js`, `ext/sse.js`, `ext/live.js`, `ext/preload.js`), Vazirmatn woff2.
5. Run `npx htmx.org@4.0.0 upgrade-check` on all templates before Done.

## 3) Repo layout
```
workshop-zero/
├── PROMPT.md  README.md (fa)  LICENSE (MIT)
├── go.mod                     # go 1.27; deps: stdlib + github.com/skip2/go-qrcode only
├── cmd/server/main.go         # flags: -addr, -data, freeze subcommand
├── internal/
│   ├── web/                   # handlers, SSE hub, fragment rendering
│   ├── quiz/                  # scoring engine + quiz-data.json (embedded)
│   ├── roadmap/               # template roadmaps + AI forge + provider chain
│   └── store/                 # append-only JSONL + in-memory index (no DB, no CGO)
├── web/
│   ├── templates/             # html/template: layouts, pages, fragments/
│   ├── static/                # vendor/ (htmx, fonts), css/, js/ (tsc output), og/
│   └── ts/                    # 3 small TS modules (see §4)
├── content/                   # fa markdown: stations, sessions, resources (embedded)
├── Dockerfile  render.yaml  .github/workflows/pages.yml
└── tsconfig.json  package.json
```

## 4) Stack rules
**Go core:** stdlib only. `embed.FS` ships templates+static+content+font inside
the binary. `CGO_ENABLED=0`. Persistence = JSONL file + memory (free-tier
ephemeral disk is fine; export button for facilitator). SSE hub with
`http.ResponseController{}.Flush()`, heartbeats every 25s, auto-reconnect.
**htmx 4:** self-hosted pinned 4.0.0. Use `:inherited` explicitly; new event
names only; `hx-status:422` to swap validation errors into the form;
`<hx-partial>` when one response updates ≥2 targets; `innerMorph` for bars/
counters; `hx-preload` on hover for session cards; `hx-action`+`hx-method` on
every form so **the quiz works with JavaScript disabled** (native POST fallback).
`hx-live` for micro-interactions (keyboard 1/2/3 answer keys, theme toggle,
confetti on result) — no Alpine, no hyperscript.
**TypeScript 7:** `typescript@^7.0.0` (option: `typescript@rc` = 7.1 RC if you
want the newest — compiler-only use is safe). `tsc` emits ES modules directly to
`web/static/js/`; no bundler, no framework. Exactly three modules:
`quiz-keys.ts`, `theme.ts`, `quiz-static.ts` (client-side scoring used ONLY in
the frozen Pages build). `strict`, `rootDir:"web/ts"`, `types:[]`.

## 5) Product spec (be creative, ship these)
- `/` — Hero «از مصرف‌کننده تا سازنده», dark gradient `#0B1020`, track colors
  `#7C3AED/#06B6D4/#F59E0B`, bento of 3 tracks, 180-minute agenda timeline,
  live counter «N نفر مسیرشان را پیدا کرده‌اند» (SSE + innerMorph), CTA to quiz.
- `/quiz` — 10 Persian questions **verbatim from PROMPT-v2 §4.4**, one fragment
  per question: `POST /quiz/{id}/answer` returns next question + `<hx-partial>`
  progress bar update. Back/forward works (htmx 4 history re-fetch). Paper
  version: print stylesheet.
- `/result/{id}` — three score bars (count/10 ×100), morph-animated; tie rule:
  gap <15% → dual-track verdict card; shareable URL + QR (go-qrcode) + SVG OG
  image generated in Go; «کامیت اولت را بزن» GitHub hello-world guide box.
- `/roadmap/{id}` — **the Roadmap Forge (signature feature):** personalized
  15-session plan (workshop 00 + chosen track(s) + Demo Day). Streams in over
  SSE as HTML fragments (skeleton → sessions appear one by one). Each session
  card expands via `hx-get` (goals, artifact, tools with ✅/🔵/🟡 layer badges),
  checkbox «انجام شد» persists to JSONL → progress bar morphs. Export as
  Markdown + Marp slides.
- `/live` — facilitator dashboard (unguessable path + rate limit): SSE aggregate
  of all quiz results, track distribution bars, station rotation timer synced
  to the 180-min agenda, «demo failed? play backup» checklist from v2 §4.3.
- `/stations` `/resources` `/faq` — station kits, three-layer access table,
  countdown to workshop date (config flag), Telegram/Porsline links.

## 6) AI Roadmap engine (free, resilient, safe)
- Provider chain, all OpenAI-compatible, keys server-side only, 20s timeout:
  **Zen → NVIDIA NIM → Zhipu GLM-Flash → Ollama (localhost, optional) → static**.
- LLM returns **strict JSON** (never raw HTML): `{track, headline_fa,
  weeks:[{n, title_fa, goal_fa, artifact_fa, tools:[{name, layer}]}]}`.
  Go validates + renders via templates → no injection surface. System prompt in
  Persian, grounded in `content/sessions/*` so it recommends REAL sessions.
- Cache by hash(quiz profile + track) in JSONL → quota-friendly on free tiers.
  Per-IP rate limit (e.g. 5 forges/hour). If every provider fails: deterministic
  template roadmap renders instantly — **the demo never dies, VPN or no VPN.**

## 7) Deploy ladder (cleanest → most public)
A. **Local:** `go run ./cmd/server` → :8080. That's the whole stack.
B. **Instant public preview (30 seconds, no account):**
   `cloudflared tunnel --url http://localhost:8080` → share the trycloudflare URL
   with the class. Ideal for the lab PC / your own machine.
C. **Free cloud (no card):** push to GitHub → Render free web service via
   `render.yaml` (Docker). Multi-stage Dockerfile:
   node:24-alpine (npm ci → tsc → vendor htmx/fonts) → golang:1.27-alpine
   (CGO_ENABLED=0 go build) → scratch. Target image < 25 MB. Note the 15-min
   sleep: add a «بیدارباش» warm-up ping note in README for demo days.
D. **Permanent Iran-friendly mirror:** `go run ./cmd/server freeze` exports a
   static snapshot (quiz scored client-side by `quiz-static.ts`, roadmaps =
   template versions) → GitHub Pages via Actions. Same templates, zero server.

## 8) Definition of Done
- [ ] `go build` → one binary; `tsc` strict clean; `upgrade-check` zero issues.
- [ ] No runtime request to any external domain (fonts/htmx vendored).
- [ ] Quiz completable with JS disabled; SSE reconnects after network drop.
- [ ] Roadmap renders with ALL API keys removed (static fallback proven).
- [ ] htmx 4 idioms only: `:inherited`, new events, `hx-status`, `<hx-partial>`.
- [ ] RTL + Persian digits (`Intl.NumberFormat('fa-IR')` / Go formatter),
      dark/light theme, `prefers-reduced-motion`, Lighthouse ≥90 perf / ≥95 a11y.
- [ ] Deployed: Quick Tunnel tested + Render live + Pages mirror live.

## 9) Sources (study list)
four.htmx.org/docs • four.htmx.org/docs/whats-new-in-htmx-4 •
devblogs.microsoft.com/typescript/announcing-typescript-7-0/ •
go.dev/doc/go1.27 • developers.cloudflare.com/tunnel/get-started/quick-tunnels/ •
render.com/docs/free • build.nvidia.com • open.bigmodel.cn • opencode.ai •
github.com/rastikerdar/vazirmatn • PROMPT-v2 (workshop content, quiz, layers)
