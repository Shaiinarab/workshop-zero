// Client-side quiz scoring for the frozen static build. Used ONLY where there
// is no server to POST answers to; the server build scores with the same rules.
export const SELECTION_SIZE = 4;
export const ALL_MODULES = ['m1', 'm2', 'm3', 'm4', 'm5', 'm6', 'm7', 'm8'];
export function scoreQuiz(questions, answers) {
    const counts = {};
    answers.forEach((answer, i) => {
        const option = questions[i]?.options.find((o) => o.key === answer);
        for (const m of option?.modules ?? [])
            counts[m] = (counts[m] ?? 0) + 1;
    });
    const ranked = Object.entries(counts).sort((a, b) => (b[1] - a[1]) || a[0].localeCompare(b[0]));
    const selected = ranked.slice(0, SELECTION_SIZE).map(([id]) => id);
    for (const id of ALL_MODULES) {
        if (selected.length >= SELECTION_SIZE)
            break;
        if (!selected.includes(id))
            selected.push(id);
    }
    selected.sort();
    return { counts, selected };
}
const escape = (s) => s.replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c]);
export function initStaticQuiz() {
    const root = document.querySelector('[data-static-quiz]');
    const dataEl = document.getElementById('quiz-data');
    if (!root || !dataEl?.textContent)
        return;
    let modules = [];
    try {
        modules = JSON.parse(root.dataset.modules ?? '[]');
    }
    catch {
        modules = [];
    }
    let questions;
    try {
        questions = JSON.parse(dataEl.textContent);
    }
    catch {
        root.innerHTML = '<div class="card">داده آزمون خراب است.</div>';
        return;
    }
    if (!Array.isArray(questions) || questions.length === 0)
        return;
    const byId = new Map(modules.map((m) => [m.id, m]));
    const answers = [];
    let index = 0;
    const renderResult = () => {
        const score = scoreQuiz(questions, answers);
        const cards = score.selected
            .map((id, i) => {
            const m = byId.get(id);
            if (!m)
                return '';
            return `<article class="card">
          <div class="row between"><span class="tag accent">ماژول ${i + 1}</span><span class="tiny muted">${escape(m.id)}</span></div>
          <h3 style="margin-top:.7rem">${escape(m.title)}</h3>
          <p class="muted small">${escape(m.tagline)}</p>
          <p class="small">${escape(m.summary)}</p>
        </article>`;
        })
            .join('');
        root.innerHTML = `
      <header style="max-width:60ch">
        <span class="tag accent">نتیجه آزمون مسیر</span>
        <h1 style="margin-top:.9rem">چهار ماژول پیشنهادی برای تو</h1>
        <p class="muted">این نسخه آفلاین است؛ امتیازدهی در مرورگر انجام شد و به سرور ارسال نشد.</p>
      </header>
      <section class="block tight"><div class="container">
        <div class="grid two">${cards}</div>
        <div class="card" style="margin-top:1.4rem">
          <div class="row between wrap">
            <div>
              <h3 style="margin:0">نقشه راه تو آماده است</h3>
              <p class="muted small" style="margin:.3rem 0 0">۱۵ نشست: کارگاه صفر، هسته مشترک، هشت نشست ماژول‌ها، استودیوی پروژه و Demo Day.</p>
            </div>
            <a class="btn accent" href="../roadmap/${score.selected.join(',')}/index.html">دیدن نقشه راه</a>
          </div>
        </div>
      </div></section>`;
    };
    const render = () => {
        if (index >= questions.length) {
            renderResult();
            return;
        }
        const q = questions[index];
        root.innerHTML = `
      <article class="card" data-static-question>
        <div class="row between"><span class="muted small">سؤال ${index + 1} از ${questions.length}</span></div>
        <div class="progress" style="margin:.9rem 0"><span style="width:${Math.round((index * 100) / questions.length)}%"></span></div>
        <h2>${escape(q.text)}</h2>
        <div class="qopts">
          ${q.options
            .map((o) => `<button type="button" data-key="${escape(o.key)}"><span class="key">${escape(o.key)}</span><span>${escape(o.text)}</span></button>`)
            .join('')}
        </div>
      </article>`;
        root.querySelectorAll('button[data-key]').forEach((btn) => {
            btn.addEventListener('click', () => {
                answers.push(btn.dataset.key);
                index += 1;
                render();
            });
        });
    };
    render();
}
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => initStaticQuiz());
}
else {
    initStaticQuiz();
}
