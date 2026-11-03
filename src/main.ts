import "./styles/base.css";
import "./styles/components.css";
import "./styles/login.css";
import "./styles/dashboard.css";
import "./styles/danmaku.css";
import { h, html, listen, swap } from "./lib/dom";
import { root } from "./lib/signal";
import { boot, bootError, phase, refreshRoom } from "./state";
import { icons, mark } from "./ui/icons";
import { escapeModal, modalOpen } from "./ui/modal";
import { mountToasts } from "./ui/toast";
import { openAbout } from "./views/about";
import { Dashboard } from "./views/dashboard";
import { copyStream, toggleLive } from "./views/live";
import { LoginView } from "./views/login";
import { toggleLog } from "./views/log";
import { focusTitle, openAreaPicker } from "./views/settings";

const app = document.querySelector<HTMLElement>("#app")!;

function Boot() {
  return h(
    "div",
    { class: "boot" },
    h("div", { class: "skel-topbar" }, h("span", { class: "skel skel-mark" }), h("span", { class: "skel skel-line w-40" }), h("span", { class: "skel skel-pill" })),
    h(
      "div",
      { class: "skel-columns" },
      h("div", { class: "skel-col" }, h("div", { class: "skel skel-card tall" }), h("div", { class: "skel-row" }, h("div", { class: "skel skel-card" }), h("div", { class: "skel skel-card" }))),
      h("div", { class: "skel skel-card side" }),
    ),
  );
}

function BootError() {
  return h(
    "div",
    { class: "login-screen" },
    h(
      "section",
      { class: "login-card error-card" },
      html(mark),
      h("h1", null, "无法连接后端"),
      h("p", { class: "muted" }, bootError()),
      h("button", { class: "btn btn-primary", onclick: () => void boot() }, html(icons.refresh), "重试"),
    ),
  );
}

root(() => {
  mountToasts(document.body);
  swap(
    app,
    phase,
    (p) => {
      if (p === "boot") return Boot();
      if (p === "unauth") return LoginView();
      if (p === "error") return BootError();
      return Dashboard();
    },
    { animate: true },
  );

  listen(window, "keydown", (e) => {
    if (e.key === "Escape") {
      if (escapeModal()) e.preventDefault();
      else if (isTyping(e.target)) (e.target as HTMLElement).blur();
      return;
    }
    if (e.ctrlKey || e.metaKey || e.altKey || e.isComposing || isTyping(e.target)) return;
    if (e.key === "?") {
      e.preventDefault();
      openAbout();
      return;
    }
    if (modalOpen.peek() > 0 || phase.peek() !== "app") return;
    const run = (fn: () => void) => {
      e.preventDefault();
      fn();
    };
    switch (e.code) {
      case "Space":
        return run(toggleLive);
      case "KeyT":
        return run(focusTitle);
      case "KeyA":
        return run(openAreaPicker);
      case "KeyR":
        return run(() => void refreshRoom());
      case "KeyC":
        return run(() => void copyStream(e.shiftKey));
      case "KeyL":
        return run(toggleLog);
    }
  });
});

function isTyping(t: EventTarget | null): boolean {
  if (!(t instanceof HTMLElement)) return false;
  if (t instanceof HTMLInputElement) return !["button", "checkbox", "radio"].includes(t.type);
  return t instanceof HTMLTextAreaElement || t.isContentEditable;
}

void boot();
