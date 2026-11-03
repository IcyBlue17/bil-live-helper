import { h, html, listen, text } from "../lib/dom";
import { effect, signal } from "../lib/signal";
import { compact, duration } from "../lib/format";
import { copy, logout, now, refreshing, refreshRoom, room, user } from "../state";
import { icons, mark } from "../ui/icons";
import { flashCopied, tween } from "../ui/widgets";
import { openAbout } from "./about";

export function statusPill() {
  return h(
    "span",
    { class: () => `status-pill status-${room()?.status ?? "offline"}` },
    h("span", { class: "status-dot" }),
    h("span", { class: "status-label" }, text(() => {
      const r = room();
      if (r?.status === "live") return "直播中";
      if (r?.status === "round") return "轮播中";
      return "未开播";
    })),
    h("span", { class: "status-time num" }, text(() => {
      const r = room();
      return r?.status === "live" ? duration(r.liveSince, now()) : "";
    })),
  );
}

export function Topbar() {
  const roomIdBtn: HTMLButtonElement = h(
    "button",
    {
      class: "room-id copy-btn",
      title: "点击复制房间号",
      onclick: async () => {
        const id = room()?.roomId;
        if (id && (await copy(String(id), "房间号"))) flashCopied(roomIdBtn);
      },
    },
    h("span", { class: "muted" }, "房间"),
    h("span", { class: "num" }, text(() => String(room()?.roomId ?? "—"))),
    h("span", { class: "copy-icons" }, html(icons.copy), html(icons.check)),
  );

  return h(
    "header",
    { class: "topbar" },
    h(
      "div",
      { class: "topbar-left" },
      html(mark),
      h(
        "div",
        { class: "topbar-room" },
        h("div", { class: "room-title", title: () => room()?.title ?? "" }, text(() => room()?.title || "我的直播间")),
        h("div", { class: "room-meta" }, roomIdBtn, h("span", { class: "sep" }), h("span", null, text(() => {
          const r = room();
          return r ? `${r.parentAreaName} / ${r.areaName}` : "";
        }))),
      ),
    ),
    h(
      "div",
      { class: "topbar-right" },
      statusPill(),
      h(
        "div",
        { class: "stat", title: "在线人数" },
        html(icons.users),
        tween(() => room()?.online ?? 0, compact),
      ),
      h(
        "div",
        { class: "stat", title: "粉丝数" },
        html(icons.heart),
        tween(() => room()?.attention ?? 0, compact),
      ),
      h(
        "button",
        {
          class: () => `icon-btn ${refreshing() ? "spinning" : ""}`,
          title: "刷新直播间信息 (R)",
          "aria-label": "刷新直播间信息",
          onclick: () => void refreshRoom(),
        },
        html(icons.refresh),
      ),
      UserMenu(),
    ),
  );
}

function avatar() {
  const box = h("span", { class: "avatar" });
  effect(() => {
    const u = user();
    if (u?.face) {
      const img = h("img", { src: u.face, alt: "", referrerpolicy: "no-referrer" });
      img.addEventListener("error", () => box.replaceChildren(initial(u.name)));
      box.replaceChildren(img);
    } else box.replaceChildren(initial(u?.name ?? ""));
  });
  return box;
}

function initial(name: string) {
  return h("span", null, [...name][0] ?? "?");
}

function UserMenu() {
  const open = signal(false);
  const wrap = h(
    "div",
    { class: () => `menu-wrap ${open() ? "open" : ""}` },
    h(
      "button",
      {
        class: "user-btn",
        "aria-haspopup": "menu",
        "aria-expanded": () => String(open()),
        onclick: () => open.update((v) => !v),
      },
      avatar(),
      h("span", { class: "user-name" }, text(() => user()?.name ?? "已登录")),
      html(icons.chevronDown),
    ),
    h(
      "div",
      { class: "menu", role: "menu" },
      h("div", { class: "menu-head" }, h("b", null, text(() => user()?.name ?? "")), h("span", { class: "muted num" }, text(() => (user() ? `UID ${user()!.uid}` : "")))),
      h("button", { role: "menuitem", onclick: () => { open.set(false); openAbout(); } }, html(icons.keyboard), "关于与快捷键", h("kbd", null, "?")),
      h("button", { role: "menuitem", class: "danger", onclick: () => { open.set(false); void logout(); } }, html(icons.logout), "退出登录"),
    ),
  );
  listen(document, "mousedown", (e) => {
    if (open.peek() && !wrap.contains(e.target as Node)) open.set(false);
  });
  listen(document, "keydown", (e) => {
    if ((e as KeyboardEvent).key === "Escape" && open.peek()) {
      open.set(false);
      e.stopPropagation();
    }
  });
  return wrap;
}
