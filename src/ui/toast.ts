import { h, html } from "../lib/dom";
import { icons } from "./icons";

export type ToastKind = "ok" | "error" | "info" | "warn";

let host: HTMLElement | null = null;

export function mountToasts(parent: HTMLElement) {
  host = h("div", { class: "toasts", role: "status", "aria-live": "polite" });
  parent.appendChild(host);
}

export function toast(message: string, kind: ToastKind = "info", ms = kind === "error" ? 5200 : 2600) {
  if (!host) return;
  const icon = kind === "ok" ? icons.check : kind === "error" ? icons.alert : kind === "warn" ? icons.alert : icons.info;
  const el = h("div", { class: `toast toast-${kind}` }, html(icon), h("span", null, message));
  host.appendChild(el);
  while (host.children.length > 4) host.firstElementChild?.remove();
  const close = () => {
    el.classList.add("leaving");
    setTimeout(() => el.remove(), 200);
  };
  el.addEventListener("click", close);
  setTimeout(close, ms);
}
