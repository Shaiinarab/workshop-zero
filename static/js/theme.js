const KEY = 'wz-theme';
export function initTheme() {
    const root = document.documentElement;
    const saved = localStorage.getItem(KEY);
    if (saved === 'light' || saved === 'dark')
        root.dataset.theme = saved;
    document.addEventListener('click', (event) => {
        const target = event.target;
        if (!target?.closest('[data-toggle-theme]'))
            return;
        const next = root.dataset.theme === 'light' ? 'dark' : 'light';
        root.dataset.theme = next;
        localStorage.setItem(KEY, next);
    });
}
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initTheme);
}
else {
    initTheme();
}
