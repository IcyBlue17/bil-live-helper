// Small reusable pieces: copy buttons, tweened numbers, status pills.
import { h, html, reducedMotion } from "../lib/dom";
import { effect, onCleanup } from "../lib/signal";
import { copy } from "../state";
import { icons } from "./icons";

/** A button that copies text and shows a check mark for a moment. */
export function copyButton(get: () => string, label: string, opts: { text?: string; class?: string } = {}) {
  const btn = h(
    "button",
    {
      class: `copy-btn ${opts.text ? "btn btn-ghost btn-sm" : "icon-btn"} ${opts.class ?? ""}`,
      title: `复制${label}`,
      "aria-label": `复制${label}`,
      onclick: async () => {
        const v = get();
        if (!v) return;
        if (await copy(v, label)) flashCopied(btn);
      },
    },
    h("span", { class: "copy-icons" }, html(icons.copy), html(icons.check)),
    opts.text ? h("span", null, opts.text) : null,
  );
  return btn;
}

const flashTimers = new WeakMap<HTMLElement, ReturnType<typeof setTimeout>>();
/** Shows a copy button's check-mark state. */
export function flashCopied(btn: HTMLElement | null) {
  if (!btn) return;
  clearTimeout(flashTimers.get(btn));
  btn.classList.add("copied");
  flashTimers.set(btn, setTimeout(() => btn.classList.remove("copied"), 1400));
}

/** A number that tweens to its new value. */
export function tween(get: () => number, format: (n: number) => string = (n) => n.toLocaleString("zh-CN")) {
  const el = h("span", { class: "num" });
  let shown = 0;
  let frame = 0;
  let first = true;
  effect(() => {
    const target = get();
    cancelAnimationFrame(frame);
    if (first || reducedMotion() || Math.abs(target - shown) < 1) {
      first = false;
      shown = target;
      el.textContent = format(target);
      return;
    }
    const from = shown;
    const start = performance.now();
    const dur = 600;
    const step = (t: number) => {
      const p = Math.min(1, (t - start) / dur);
      const e = 1 - Math.pow(1 - p, 3);
      shown = from + (target - from) * e;
      el.textContent = format(Math.round(shown));
      if (p < 1) frame = requestAnimationFrame(step);
    };
    frame = requestAnimationFrame(step);
  });
  onCleanup(() => cancelAnimationFrame(frame));
  return el;
}

export function kbd(...keys: string[]) {
  return h("span", { class: "kbds" }, keys.map((k) => h("kbd", null, k)));
}

export function spinner() {
  return h("span", { class: "spinner", "aria-hidden": "true" });
}
