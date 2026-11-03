import { h, html, reducedMotion } from "../lib/dom";
import { root, signal } from "../lib/signal";
import { icons } from "./icons";

type Open = { close: () => void; onEsc: () => void };
const stack: Open[] = [];
export const modalOpen = signal(0);

/** Closes the top modal; returns whether there was one. */
export function escapeModal(): boolean {
  const top = stack[stack.length - 1];
  if (!top) return false;
  top.onEsc();
  return true;
}

export interface ModalOptions {
  title: string;
  subtitle?: string;
  width?: number;
  class?: string;
  /** Called on Esc, backdrop click or the close button; defaults to close. */
  onDismiss?: () => void;
}

/** Opens a modal whose body view runs in its own scope; returns a closer. */
export function openModal(opts: ModalOptions, body: (close: () => void) => Node): () => void {
  let closed = false;
  let dispose = () => {};
  const prevFocus = document.activeElement as HTMLElement | null;
  const close = () => {
    if (closed) return;
    closed = true;
    const i = stack.indexOf(entry);
    if (i >= 0) stack.splice(i, 1);
    modalOpen.set(stack.length);
    dispose();
    backdrop.classList.add("leaving");
    setTimeout(() => backdrop.remove(), reducedMotion() ? 0 : 200);
    prevFocus?.focus?.();
  };
  const dismiss = () => (opts.onDismiss ? opts.onDismiss() : close());
  const entry: Open = { close, onEsc: dismiss };

  const [content, d] = root(() => body(close));
  dispose = d;
  const dialog = h(
    "div",
    {
      class: `modal ${opts.class ?? ""}`,
      role: "dialog",
      "aria-modal": "true",
      "aria-label": opts.title,
      style: opts.width ? `width:min(${opts.width}px, calc(100vw - 48px))` : undefined,
      tabindex: "-1",
    },
    h(
      "header",
      { class: "modal-head" },
      h("div", null, h("h2", null, opts.title), opts.subtitle ? h("p", { class: "muted" }, opts.subtitle) : null),
      h("button", { class: "icon-btn", "aria-label": "关闭", title: "关闭 (Esc)", onclick: dismiss }, html(icons.close)),
    ),
    content,
  );
  const backdrop = h("div", { class: "backdrop" }, dialog);
  backdrop.addEventListener("mousedown", (e) => {
    if (e.target === backdrop) dismiss();
  });
  document.body.appendChild(backdrop);
  stack.push(entry);
  modalOpen.set(stack.length);
  requestAnimationFrame(() => {
    const auto = dialog.querySelector<HTMLElement>("[autofocus]");
    (auto ?? dialog).focus();
  });
  return close;
}
