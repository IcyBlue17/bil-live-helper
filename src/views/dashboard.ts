// The logged-in window: top bar, the control column, the danmaku column.
import { h, html, text } from "../lib/dom";
import { onCleanup, signal } from "../lib/signal";
import { feedListeners, type FeedItem } from "../state";
import { icons } from "../ui/icons";
import { DanmakuPanel } from "./danmaku";
import { FaceAuthWatcher, LiveCard, StreamCard } from "./live";
import { LogDrawer } from "./log";
import { SettingsCard } from "./settings";
import { Topbar } from "./topbar";

/** Which column shows in the narrow (tabbed) layout. */
export const narrowTab = signal<"control" | "danmaku">("control");

export function Dashboard() {
  FaceAuthWatcher();
  // Live messages that arrived while the 弹幕 tab was hidden (narrow layout only).
  const unseenDanmaku = signal(0);
  const count = (items: FeedItem[], reset: boolean) => {
    if (reset) unseenDanmaku.set(0);
    else if (narrowTab.peek() !== "danmaku") unseenDanmaku.update((n) => n + items.filter((m) => !m.historical).length);
  };
  feedListeners.add(count);
  onCleanup(() => feedListeners.delete(count));
  const tab = (id: "control" | "danmaku", label: string, icon: string) =>
    h(
      "button",
      {
        class: () => `tab ${narrowTab() === id ? "active" : ""}`,
        role: "tab",
        "aria-selected": () => String(narrowTab() === id),
        onclick: () => {
          narrowTab.set(id);
          if (id === "danmaku") unseenDanmaku.set(0);
        },
      },
      html(icon),
      label,
      id === "danmaku" ? h("span", { class: () => `tab-badge num ${unseenDanmaku() > 0 ? "show" : ""}` }, text(() => (unseenDanmaku() > 99 ? "99+" : String(unseenDanmaku())))) : null,
    );

  return h(
    "div",
    { class: () => `dashboard tab-${narrowTab()}` },
    Topbar(),
    h("nav", { class: "tabs", role: "tablist" }, tab("control", "直播控制", icons.grid), tab("danmaku", "弹幕", icons.chat)),
    h(
      "main",
      { class: "columns" },
      h("div", { class: "col col-control" }, LiveCard(), h("div", { class: "row-2" }, StreamCard(), SettingsCard())),
      h("div", { class: "col col-danmaku" }, DanmakuPanel()),
    ),
    LogDrawer(),
  );
}
