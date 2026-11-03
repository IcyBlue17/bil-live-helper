// The 弹幕 panel: connection state, the feed, and the send box.
import { h, html, reducedMotion, text } from "../lib/dom";
import { computed, onCleanup, signal } from "../lib/signal";
import { charCount, clock, compact, errorMessage, yuan } from "../lib/format";
import { MAX_FEED, conn, feed, feedListeners, phase, popularity, sendDanmaku, user, type FeedItem } from "../state";
import type { ConnState, Emoticon, GuardLevel } from "../mygo";
import { icons } from "../ui/icons";
import { spinner, tween } from "../ui/widgets";

export const DANMAKU_MAX = 40;

const connLabel: Record<ConnState, string> = {
  connecting: "连接中",
  connected: "已连接",
  reconnecting: "重连中",
  closed: "未连接",
};

const guardName: Record<GuardLevel, string> = { 0: "", 1: "总督", 2: "提督", 3: "舰长" };

export function DanmakuPanel() {
  const list = h("div", { class: "feed-list", role: "log", "aria-live": "off" });
  const scroller = h("div", { class: "feed-scroll" }, list);
  const unseen = signal(0);
  let atBottom = true;
  let historyDivider: HTMLElement | null = null;

  const nearBottom = () => scroller.scrollHeight - scroller.scrollTop - scroller.clientHeight < 32;
  const toBottom = (smooth: boolean) => {
    scroller.scrollTo({ top: scroller.scrollHeight, behavior: smooth && !reducedMotion() ? "smooth" : "auto" });
    unseen.set(0);
  };
  scroller.addEventListener("scroll", () => {
    atBottom = nearBottom();
    if (atBottom) unseen.set(0);
  });

  const empty = h(
    "div",
    { class: "empty feed-empty" },
    h("span", { class: "empty-icon" }, html(icons.chat)),
    h("p", null, "暂无弹幕"),
    h("p", { class: "muted" }, "观众发送的弹幕、礼物和进场消息会显示在这里。"),
  );

  const render = (items: FeedItem[], reset: boolean) => {
    if (reset) {
      list.replaceChildren();
      historyDivider = null;
      atBottom = true;
      unseen.set(0);
    }
    const self = user.peek()?.uid;
    for (const m of items) {
      // The divider goes between the last history item and the first live one.
      if (!m.historical && !historyDivider && list.querySelector(".msg.historical")) {
        historyDivider = h("div", { class: "feed-divider" }, h("span", null, "以上为历史弹幕"));
        list.appendChild(historyDivider);
      }
      list.appendChild(messageRow(m, self));
    }
    // Cap the DOM like the feed signal.
    let extra = list.querySelectorAll(".msg").length - MAX_FEED;
    while (extra > 0 && list.firstElementChild) {
      const first = list.firstElementChild;
      if (first.classList.contains("msg")) extra--;
      if (first === historyDivider) historyDivider = null;
      first.remove();
    }
    empty.classList.toggle("hidden", list.childElementCount > 0);
    if (atBottom) requestAnimationFrame(() => toBottom(false));
    else unseen.update((n) => n + items.filter((m) => !m.historical).length);
  };
  feedListeners.add(render);
  onCleanup(() => feedListeners.delete(render));
  render(feed.peek(), true);

  const newPill = h(
    "button",
    { class: () => `new-pill ${unseen() > 0 ? "show" : ""}`, onclick: () => toBottom(true) },
    text(() => `${unseen() > 99 ? "99+" : unseen()} 条新弹幕`),
    html(icons.arrowDown),
  );

  return h(
    "section",
    { class: "card danmaku-panel" },
    h(
      "header",
      { class: "card-head" },
      h("h3", null, html(icons.chat), "弹幕"),
      h(
        "div",
        { class: "card-tools" },
        h("span", { class: "stat stat-sm", title: "人气值" }, html(icons.flame), tween(popularity, (n) => (n ? compact(n) : "—"))),
        h("span", { class: () => `conn-pill conn-${conn()}` }, h("span", { class: "conn-dot" }), text(() => connLabel[conn()])),
      ),
    ),
    h("div", { class: "feed" }, scroller, empty, newPill),
    SendBox(),
  );
}

function messageRow(m: FeedItem, self: number | undefined): HTMLElement {
  const cls = ["msg", `msg-${m.kind}`];
  if (m.historical) cls.push("historical");
  if (m.fresh && !reducedMotion()) cls.push("fresh");
  const isSelf = m.kind === "self" || (self !== undefined && m.uid === self && m.kind === "danmaku");
  const time = h("time", { class: "msg-time num", datetime: m.time }, clock(m.time));

  switch (m.kind) {
    case "superchat": {
      const sc = m.superChat;
      return h(
        "div",
        { class: cls.join(" ") },
        h(
          "div",
          { class: "sc-card" },
          h(
            "div",
            { class: "sc-head" },
            m.medal ? h("span", { class: "medal" }, h("span", null, m.medal.name), h("b", { class: "num" }, String(m.medal.level))) : null,
            h("span", { class: "msg-name" }, m.name),
            h("span", { class: "sc-price num" }, sc ? `¥${sc.price}` : "SC"),
          ),
          h("div", { class: "sc-text" }, m.text),
        ),
      );
    }
    case "follow":
    case "like":
      return h(
        "div",
        { class: cls.join(" ") },
        time,
        h("div", { class: "msg-body" }, h("span", { class: "msg-name" }, m.name), m.text || (m.kind === "follow" ? " 关注了直播间" : " 点赞了直播间")),
      );
    case "guard":
    case "gift": {
      const g = m.gift;
      if (m.guard) cls.push(`guard-${m.guard}`);
      else if (g && g.price >= 50_000) cls.push("gift-big");
      return h(
        "div",
        { class: cls.join(" ") },
        time,
        h(
          "div",
          { class: "msg-body" },
          html(icons.gift),
          h("span", { class: "msg-name" }, m.name),
          h("span", { class: "muted" }, m.kind === "guard" || m.guard ? "开通了" : "送出"),
          h("span", { class: "gift-name" }, g ? `${g.name}${g.num > 1 ? ` ×${g.num}` : ""}` : m.text),
          g && g.price > 0 ? h("span", { class: "gift-price num" }, yuan(g.price)) : null,
        ),
      );
    }
    case "enter":
      return h("div", { class: cls.join(" ") }, time, h("div", { class: "msg-body" }, h("span", { class: "msg-name" }, m.name), m.text || " 进入直播间"));
    case "notice":
    case "system":
      return h("div", { class: cls.join(" ") }, h("div", { class: "msg-body" }, m.text));
    default: {
      if (isSelf) cls.push("msg-self");
      if (m.guard) cls.push(`guard-${m.guard}`);
      else if (m.medal) cls.push("fan");
      if (m.admin) cls.push("admin");
      return h(
        "div",
        { class: cls.join(" ") },
        time,
        h(
          "div",
          { class: "msg-body" },
          h(
            "span",
            { class: "msg-who" },
            m.admin ? h("span", { class: "tag tag-admin", title: "房管" }, "房") : null,
            m.guard ? h("span", { class: `tag tag-guard tag-guard-${m.guard}` }, guardName[m.guard]) : null,
            m.medal ? h("span", { class: "medal" }, h("span", null, m.medal.name), h("b", { class: "num" }, String(m.medal.level))) : null,
            m.userLevel ? h("span", { class: "ul num", title: "用户等级" }, `UL${m.userLevel}`) : null,
            h("span", { class: "msg-name" }, isSelf ? `${m.name}（我）` : m.name),
          ),
          messageText(m),
        ),
      );
    }
  }
}

/** The text of a danmaku: a sticker, or text with inline emotes. */
function messageText(m: FeedItem): HTMLElement {
  if (m.emoticon) return h("span", { class: "msg-text" }, emote(m.emoticon, true));
  const tokens = Object.keys(m.emots ?? {});
  if (!tokens.length) return h("span", { class: "msg-text" }, m.text);
  const re = new RegExp(tokens.map((t) => t.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")).join("|"), "g");
  const parts: (Node | string)[] = [];
  let last = 0;
  for (const match of m.text.matchAll(re)) {
    const e = m.emots[match[0]];
    if (!e) continue;
    parts.push(m.text.slice(last, match.index));
    parts.push(emote(e, false));
    last = match.index + match[0].length;
  }
  parts.push(m.text.slice(last));
  return h("span", { class: "msg-text" }, parts);
}

function emote(e: Emoticon, sticker: boolean): HTMLElement {
  return h("img", {
    class: sticker ? "emote-sticker" : "emote",
    src: e.url,
    alt: e.name,
    title: e.name,
    referrerpolicy: "no-referrer",
    loading: "lazy",
  });
}

function SendBox() {
  const value = signal("");
  const sending = signal(false);
  const error = signal("");
  const count = computed(() => charCount(value()));
  const canSend = computed(() => value().trim() !== "" && count() <= DANMAKU_MAX && !sending() && phase() === "app");

  const send = async () => {
    if (!canSend()) {
      if (value().trim() === "") error.set("弹幕内容不能为空");
      return;
    }
    sending.set(true);
    error.set("");
    const text = value().trim();
    try {
      await sendDanmaku(text);
      value.set("");
    } catch (err) {
      error.set(`发送失败：${errorMessage(err)}`);
    } finally {
      sending.set(false);
      input.focus();
    }
  };

  const input = h("input", {
    class: () => `input send-input ${count() > DANMAKU_MAX ? "invalid" : ""}`,
    id: "danmaku-input",
    type: "text",
    placeholder: "发送弹幕…",
    autocomplete: "off",
    "aria-label": "弹幕内容",
    value: () => value(),
    oninput: (e: Event) => {
      value.set((e.target as HTMLInputElement).value);
      if (error()) error.set("");
    },
    onkeydown: (e: KeyboardEvent) => {
      if (e.key === "Enter" && !e.isComposing) {
        e.preventDefault();
        void send();
      }
    },
  });

  return h(
    "footer",
    { class: "send-box" },
    h(
      "div",
      { class: "send-row" },
      h("div", { class: "send-field" }, input, h("span", { class: () => `counter num ${count() > DANMAKU_MAX ? "over" : count() > DANMAKU_MAX - 5 ? "near" : ""}` }, text(() => `${count()}/${DANMAKU_MAX}`))),
      h(
        "button",
        { class: "btn btn-primary send-btn", disabled: () => !canSend(), onclick: () => void send(), title: "发送 (Enter)", "aria-label": "发送弹幕" },
        h("span", { class: () => (sending() ? "" : "hidden") }, spinner()),
        h("span", { class: () => (sending() ? "hidden" : "") }, html(icons.send)),
      ),
    ),
    h("p", { class: () => `send-error ${error() ? "show" : ""}`, role: "alert" }, html(icons.alert), text(() => error())),
  );
}
