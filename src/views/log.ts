// The collapsible activity log drawer at the bottom of the dashboard.
import { h, html, text } from "../lib/dom";
import { effect } from "../lib/signal";
import { clock } from "../lib/format";
import { logOpen, logs, unreadLogs } from "../state";
import { icons } from "../ui/icons";

const levelLabel = { info: "信息", ok: "成功", warn: "警告", error: "错误" } as const;

export function toggleLog() {
  logOpen.update((v) => !v);
}

export function LogDrawer() {
  const list = h("div", { class: "log-list", role: "log" });
  effect(() => {
    list.replaceChildren(
      ...logs().map((e) =>
        h("div", { class: `log-line log-${e.level}` }, h("time", { class: "num" }, clock(e.time)), h("span", { class: "log-level" }, levelLabel[e.level]), h("span", { class: "log-msg" }, e.message)),
      ),
    );
    if (logOpen.peek()) requestAnimationFrame(() => (list.scrollTop = list.scrollHeight));
  });
  effect(() => {
    if (logOpen()) {
      unreadLogs.set(0);
      requestAnimationFrame(() => (list.scrollTop = list.scrollHeight));
    }
  });

  const last = () => logs()[logs().length - 1];
  return h(
    "section",
    { class: () => `log-drawer ${logOpen() ? "open" : ""}` },
    h(
      "button",
      { class: "log-bar", onclick: toggleLog, "aria-expanded": () => String(logOpen()), title: "活动日志 (L)" },
      h("span", { class: "log-title" }, html(icons.terminal), "活动日志"),
      h("span", { class: () => `log-unread num ${unreadLogs() > 0 && !logOpen() ? "show" : ""}` }, text(() => String(unreadLogs()))),
      h("span", { class: () => `log-last log-${last()?.level ?? "info"}` }, text(() => {
        const l = last();
        return l ? `${clock(l.time)}  ${l.message}` : "暂无记录";
      })),
      h("kbd", null, "L"),
      h("span", { class: "log-chevron" }, html(icons.chevronDown)),
    ),
    h("div", { class: "log-body" }, list),
  );
}
