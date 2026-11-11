// Keyboard answering: 1/2/3 map to the first three options of the question that
// is currently on screen. Delegation on `document` means it keeps working after
// htmx swaps a new question in, with no re-binding hook.

const KEY_TO_OPTION: Record<string, number> = { '1': 0, '2': 1, '3': 2 };

export function initQuizKeys(): void {
  document.addEventListener('keydown', (event) => {
    if (event.defaultPrevented || event.altKey || event.ctrlKey || event.metaKey) return;
    const el = event.target as HTMLElement | null;
    if (el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.isContentEditable)) return;

    const index = KEY_TO_OPTION[event.key];
    if (index === undefined) return;

    const form = document.querySelector<HTMLFormElement>('[data-quiz-form]');
    if (!form) return;

    const buttons = form.querySelectorAll<HTMLButtonElement>('button[data-answer]');
    const button = buttons[index];
    if (!button) return;

    event.preventDefault();
    button.click();
  });
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initQuizKeys);
} else {
  initQuizKeys();
}